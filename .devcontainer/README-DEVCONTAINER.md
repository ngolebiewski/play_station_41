# Dev Container Configuration (CI/CD & Headless Tooling)

    💡 Architectural Note: This configuration is maintained as an isolated infrastructure branch to provide a deterministic, containerized environment for automated headless testing and CI/CD consistency. For day-to-day active development on Apple Silicon, native execution is preferred to directly leverage macOS hardware acceleration (Metal API for graphics, CoreAudio for sound) without the performance overhead of graphical emulation inside a Linux container.

## How to Initialize the Dev Container in VS Code

### If you need to validate the build environment or run automated scripts inside the standardized container, follow these steps:

    1. Start Docker: Ensure Docker Desktop is launched and running on your host machine.

    2. Install Extension: Open VS Code and verify the official Microsoft Dev Containers extension is installed.

    3. Open Project: Open the play_station_41 root folder in VS Code.

    4. Reopen in Container:

        * Open the Command Palette (Cmd + Shift + P on Mac, Ctrl + Shift + P on Windows/Linux).

        * Select Dev Containers: Reopen in Container.

### Once VS Code reloads, a green status indicator will appear in the bottom-left corner (e.g., Dev Container: PlayStation41...), confirming that your workspace environment is fully isolated and operational inside the containerized Linux runtime.