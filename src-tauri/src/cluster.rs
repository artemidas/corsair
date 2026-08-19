//! Cluster connection handling.
//!
//! MVP scope: connect via the system's default kubeconfig. When the caller
//! supplies a context name, that context is selected from the kubeconfig
//! (resolved against the same env the default config uses — `$KUBECONFIG`
//! or `~/.kube/config`). When no name is supplied, the currently active
//! context is used, matching what `kubectl` does by default.

use kube::config::{KubeConfigOptions, Kubeconfig};
use kube::{Client, Config};
use serde::Serialize;

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ClusterStatus {
    pub connected: bool,
    pub healthy: bool,
    pub context: Option<String>,
    pub error: Option<String>,
}

impl ClusterStatus {
    pub fn disconnected() -> Self {
        Self {
            connected: false,
            healthy: false,
            context: None,
            error: None,
        }
    }

    pub fn connected(context: Option<String>) -> Self {
        Self {
            connected: true,
            healthy: true,
            context,
            error: None,
        }
    }

    pub fn unreachable(context: Option<String>, error: String) -> Self {
        Self {
            connected: true,
            healthy: false,
            context,
            error: Some(error),
        }
    }
}

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct KubeContexts {
    pub current: Option<String>,
    pub contexts: Vec<String>,
}

pub fn list_contexts() -> Result<KubeContexts, String> {
    let config = Kubeconfig::read().map_err(|e| e.to_string())?;
    let mut contexts: Vec<String> = config.contexts.iter().map(|c| c.name.clone()).collect();
    contexts.sort();
    Ok(KubeContexts {
        current: config.current_context,
        contexts,
    })
}

/// Resolve the kubeconfig context name that a connect attempt will use.
pub fn resolve_context_name(requested: Option<&str>) -> Option<String> {
    match requested {
        Some(name) if !name.is_empty() => Some(name.to_string()),
        _ => Kubeconfig::read().ok().and_then(|cfg| cfg.current_context),
    }
}

pub async fn connect(context: Option<&str>) -> Result<Client, kube::Error> {
    match context {
        Some(name) if !name.is_empty() => {
            let opts = KubeConfigOptions {
                context: Some(name.to_string()),
                ..Default::default()
            };
            let config = Config::from_kubeconfig(&opts).await?;
            Client::try_from(config)
        }
        _ => Client::try_default().await,
    }
}

/// Cheap liveness check against the apiserver (`GET /version`).
pub async fn probe(client: &Client) -> Result<(), kube::Error> {
    client.apiserver_version().await.map(|_| ())
}
