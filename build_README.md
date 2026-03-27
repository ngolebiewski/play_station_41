Commit your code changes:
```Bash

git add .
git commit -m "added a new level"
```

Bump the version:
```Bash

make bump
```

This creates a new tag (e.g., v0.1.5) in your local Git history.

Push the code AND the tags to GitHub:
```Bash

git push origin main
git push --tags
```
The --tags flag is critical; without it, GitHub won't know the version exists.

Build and Release:
```Bash

make release
```

This compiles everything and uploads the binaries to the GitHub "Releases" page.



# Starting at a specific version

If you feel like the game is further along and want to start at v0.1.0 instead of v0.0.1, you can skip the make bump for the very first time and manually set the tag:
```Bash

git tag v0.1.0
git push --tags
make release
```