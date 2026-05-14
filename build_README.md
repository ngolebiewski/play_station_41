## Commit your code changes:
```Bash

git add .
git commit -m "added a new level"
```

## Bump the version:
Choose the command that fits the scale of your update:

### Major update (e.g., v0.1.2 → v1.0.0):
```Bash

make bump-major
```

### Minor update (e.g., v0.1.2 → v0.2.0):
```Bash

make bump-minor
```

### Patch/Bugfix (e.g., v0.1.2 → v0.1.3):
```Bash

make bump-patch
```
This creates a new tag (e.g., v0.1.5) in your local Git history.

## Push the code AND the tags to GitHub:
```Bash

git push origin main
git push --tags
```
The --tags flag is critical; without it, GitHub won't know the version exists.

## Build and Release:
```Bash

make release
```

This compiles everything and uploads the binaries to the GitHub "Releases" page.



## Alternatively, Starting at a specific version

If you feel like the game is further along and want to start at v0.1.0 instead of v0.0.1, you can skip the make bump for the very first time and manually set the tag:
```Bash

git tag v0.1.0
git push --tags
make release
```

## A Quick Note on Requirements

To use the make release command, ensure you have the GitHub CLI installed and authenticated:

    Install: brew install gh (on macOS)

    Login: gh auth login