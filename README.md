# homelab

# Remote Development Setup

This project is configured for remote development using VS Code and the Remote - SSH extension. Code is edited locally but lives and runs on your Linux machine.

## Getting Started

1. Install the [Remote - SSH extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-ssh) in VS Code.
2. Ensure you have SSH access to your Linux machine.
3. Update `.vscode/settings.json` if your username is not `jweber`:
	- Change `/home/jweber/projects/homelab` to your actual home directory path.
4. Open the folder in VS Code using "Remote-SSH: Connect to Host...".

## Recommended Extensions

Extensions are listed in `.vscode/extensions.json`. VS Code will prompt you to install them when you open the project.

## Notes

- Each user should update the workspace path in `.vscode/settings.json` to match their Linux username if different.

## PreRequisites

```bash
$ sudo dnf install ansible-core ansible-collection-ansible-posix
```
