# release.ps1

# Ensure we are in the correct directory (optional, assumes script is run from repo root)
# cd $PSScriptRoot

# Fetch the latest tags from the remote
git fetch --tags origin

# Get the latest tag matching the vX.Y.Z pattern
$latestTag = git tag --sort=-v:refname | Select-Object -First 1 | Where-Object { $_ -match '^v\d+\.\d+\.\d+$' }

if (-not $latestTag) {
    Write-Host "No existing vX.Y.Z tag found. Starting with v0.1.0."
    $newTag = "v0.1.0"
} else {
    Write-Host "Latest tag found: $latestTag"
    # Split the tag into parts (v, Major, Minor, Patch)
    $versionParts = $latestTag -split '[v.]'
    # Increment the patch version
    $major = [int]$versionParts[1]
    $minor = [int]$versionParts[2]
    $patch = [int]$versionParts[3] + 1
    $newTag = "v$major.$minor.$patch"
}

Write-Host "Creating new tag: $newTag"

# Optional: Add and commit any pending changes before tagging
# git add .
# git commit -m "chore: prepare release $newTag"
# git push origin <your-default-branch> # Replace <your-default-branch> e.g., main or master

# Create the new tag
git tag $newTag

# Push the new tag to the remote
Write-Host "Pushing tag $newTag to origin..."
git push origin $newTag

Write-Host "Release process completed for tag $newTag."