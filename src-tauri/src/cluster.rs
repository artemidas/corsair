//! Cluster connection handling.
//!
//! MVP scope: connect via the system's default kubeconfig. When the caller
//! supplies a context name, that context is selected from the kubeconfig
//! (resolved against the same env the default config uses — `$KUBECONFIG`
//! or `~/.kube/config`). When no name is supplied, the currently active
//! context is used, matching what `kubectl` does by default.

use kube::config::KubeConfigOptions;
use kube::{Client, Config};

pub async fn connect(context: Option<&str>) -> Result<Client, kube::Error> {
    match context {
        Some(name) => {
            let opts = KubeConfigOptions {
                context: Some(name.to_string()),
                ..Default::default()
            };
            let config = Config::from_kubeconfig(&opts).await?;
            Client::try_from(config)
        }
        None => Client::try_default().await,
    }
}
