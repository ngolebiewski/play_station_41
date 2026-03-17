#!/bin/bash
set -e

# Navigate to repo
cd /workspaces/play_station_41

# Create new branch based on main
git fetch origin main
git checkout -b feature/character-selection origin/main

# Stage all changes
git add .

# Commit with descriptive message
git commit -m "feat: implement character selection scene (issue #4)

- Create character_selection_scene.go with grid-based UI
- Add character navigation, selection highlighting with blue outline
- Implement '1P' indicator and random selection
- Add 30-second inactivity auto-select timeout
- Update scene transition: Title -> Character Selection -> Classroom
- Add characterIndex field to Player struct
- Update title_scene.go to transition to character selection"

# Push to remote
git push -u origin feature/character-selection

echo "✓ Branch 'feature/character-selection' created and pushed successfully"
git log --oneline -1
git branch -v
