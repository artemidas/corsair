//! Cluster connection handling.
//!
//! MVP scope: connect only via the system's default kubeconfig, using
//! whatever context is currently active (same resolution `kubectl` uses:
//! `$KUBECONFIG` or `~/.kube/config`, falling back to in-cluster config).
//! No context selector yet — that's a follow-up iteration.

use kube::Client;

/// Build a `kube::Client` from the default kubeconfig / active context.
pub async fn connect() -> Result<Client, kube::Error> {
    Client::try_default().await
}
