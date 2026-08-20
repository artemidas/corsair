//! Local container image discovery via the host's `docker` or `podman` CLI.
//!
//! Tries Docker first, then Podman. Spawning the already-installed runtime is
//! not a bundled sidecar — the binaries live on the user's PATH.

use serde::{Deserialize, Serialize};
use std::env;
use std::path::{Path, PathBuf};
use std::process::Stdio;
use std::time::Duration;
use tokio::process::Command;

const LIST_TIMEOUT: Duration = Duration::from_secs(10);

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LocalImage {
    pub id: String,
    pub reference: String,
    pub repository: String,
    pub tag: String,
    pub size: String,
}

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct LocalImageList {
    pub runtime: String,
    pub images: Vec<LocalImage>,
}

#[derive(Debug, Deserialize)]
struct RawImage {
    #[serde(default, alias = "Repository")]
    repository: Option<String>,
    #[serde(default, alias = "Tag")]
    tag: Option<String>,
    #[serde(default, alias = "ID", alias = "Id")]
    id: Option<String>,
    #[serde(default, alias = "Size")]
    size: Option<serde_json::Value>,
    #[serde(default, alias = "Names")]
    names: Option<Vec<String>>,
    #[serde(default, alias = "RepoTags")]
    repo_tags: Option<Vec<String>>,
}

pub async fn list_local() -> Result<LocalImageList, String> {
    let mut errors = Vec::new();

    for runtime in ["docker", "podman"] {
        let Some(bin) = find_binary(runtime) else {
            errors.push(format!("{runtime} not found"));
            continue;
        };
        match list_with(&bin).await {
            Ok(images) => {
                return Ok(LocalImageList {
                    runtime: runtime.to_string(),
                    images,
                });
            }
            Err(e) => errors.push(format!("{runtime}: {e}")),
        }
    }

    Err(format!(
        "Could not list container images (tried docker, then podman). {}",
        errors.join(". ")
    ))
}

async fn list_with(bin: &Path) -> Result<Vec<LocalImage>, String> {
    let stdout = run(bin, &["images", "--format", "{{json .}}"]).await?;
    let mut images = parse_images(&stdout);
    images.sort_by(|a, b| a.reference.cmp(&b.reference));
    images.dedup_by(|a, b| a.reference == b.reference);
    Ok(images)
}

async fn run(bin: &Path, args: &[&str]) -> Result<String, String> {
    let mut cmd = Command::new(bin);
    cmd.args(args)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .kill_on_drop(true);

    #[cfg(windows)]
    {
        const CREATE_NO_WINDOW: u32 = 0x0800_0000;
        cmd.creation_flags(CREATE_NO_WINDOW);
    }

    let output = tokio::time::timeout(LIST_TIMEOUT, cmd.output())
        .await
        .map_err(|_| format!("timed out running {}", bin.display()))?
        .map_err(|e| format!("failed to run {}: {e}", bin.display()))?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        let stdout = String::from_utf8_lossy(&output.stdout);
        let msg = if !stderr.trim().is_empty() {
            stderr
        } else {
            stdout
        };
        return Err(msg.trim().to_string());
    }

    Ok(String::from_utf8_lossy(&output.stdout).into_owned())
}

fn parse_images(stdout: &str) -> Vec<LocalImage> {
    let trimmed = stdout.trim();
    if trimmed.is_empty() {
        return Vec::new();
    }

    let raw: Vec<RawImage> = if trimmed.starts_with('[') {
        serde_json::from_str(trimmed).unwrap_or_default()
    } else {
        trimmed
            .lines()
            .filter_map(|line| {
                let line = line.trim();
                if line.is_empty() {
                    None
                } else {
                    serde_json::from_str(line).ok()
                }
            })
            .collect()
    };

    raw.into_iter().filter_map(into_local).collect()
}

fn into_local(raw: RawImage) -> Option<LocalImage> {
    let repository = raw.repository.unwrap_or_default();
    let tag = raw.tag.unwrap_or_default();
    let reference = named_ref(&raw.names)
        .or_else(|| named_ref(&raw.repo_tags))
        .or_else(|| compose_ref(&repository, &tag))?;

    Some(LocalImage {
        id: raw.id.unwrap_or_default(),
        reference,
        repository,
        tag,
        size: format_size(raw.size.as_ref()),
    })
}

fn named_ref(names: &Option<Vec<String>>) -> Option<String> {
    names.as_ref()?.iter().find_map(|name| {
        let name = name.trim();
        if noneish(name) {
            None
        } else {
            Some(name.to_string())
        }
    })
}

fn compose_ref(repository: &str, tag: &str) -> Option<String> {
    if noneish(repository) {
        return None;
    }
    if noneish(tag) {
        Some(repository.to_string())
    } else {
        Some(format!("{repository}:{tag}"))
    }
}

fn noneish(s: &str) -> bool {
    s.is_empty() || s == "<none>"
}

fn format_size(value: Option<&serde_json::Value>) -> String {
    match value {
        Some(serde_json::Value::String(s)) => s.clone(),
        Some(serde_json::Value::Number(n)) => n
            .as_u64()
            .map(format_bytes)
            .unwrap_or_else(|| n.to_string()),
        _ => String::new(),
    }
}

fn format_bytes(n: u64) -> String {
    const KB: f64 = 1024.0;
    const MB: f64 = 1024.0 * KB;
    const GB: f64 = 1024.0 * MB;
    let n = n as f64;
    if n >= GB {
        format!("{:.1} GB", n / GB)
    } else if n >= MB {
        format!("{:.1} MB", n / MB)
    } else if n >= KB {
        format!("{:.1} KB", n / KB)
    } else {
        format!("{n:.0} B")
    }
}

fn find_binary(name: &str) -> Option<PathBuf> {
    let exe = if cfg!(windows) {
        format!("{name}.exe")
    } else {
        name.to_string()
    };

    extra_bin_dirs()
        .into_iter()
        .chain(env::split_paths(&env::var_os("PATH").unwrap_or_default()))
        .map(|dir| dir.join(&exe))
        .find(|path| path.is_file())
}

fn extra_bin_dirs() -> Vec<PathBuf> {
    let mut dirs = vec![
        PathBuf::from("/opt/homebrew/bin"),
        PathBuf::from("/usr/local/bin"),
        PathBuf::from("/opt/podman/bin"),
        PathBuf::from("/usr/bin"),
        PathBuf::from("/snap/bin"),
    ];

    if let Some(home) = env::var_os("HOME") {
        let home = PathBuf::from(home);
        dirs.push(home.join(".local/bin"));
        dirs.push(home.join(".docker/bin"));
    }

    #[cfg(windows)]
    {
        if let Some(pf) = env::var_os("ProgramFiles") {
            dirs.push(PathBuf::from(pf).join(r"Docker\Docker\resources\bin"));
        }
        if let Some(pf86) = env::var_os("ProgramFiles(x86)") {
            dirs.push(PathBuf::from(pf86).join(r"Docker\Docker\resources\bin"));
        }
    }

    dirs
}
