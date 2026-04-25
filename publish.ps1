<#
publish.ps1 - initialize repo, commit, and optionally create GitHub repo via `gh`.
Usage: .\publish.ps1 [-RepoName <name>] [-Public]
#>
param(
  [string]$RepoName = "",
  [switch]$Public
)

Set-Location -Path (Split-Path -Path $MyInvocation.MyCommand.Definition -Parent)

if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
  Write-Error "git is not installed or not in PATH. Install git first: https://git-scm.com/downloads"
  exit 1
}

# initialize if needed
if (-not (Test-Path .git)) {
  git init
  Write-Output "Initialized empty git repository."
} else {
  Write-Output "Git repository already exists."
}

# ensure .gitignore present
if (-not (Test-Path .gitignore)) {
  @"# ignore build artifacts
*.exe
*~
"@ > .gitignore
  git add .gitignore
}

# add and commit
git add -A
$commit = git rev-parse --verify HEAD 2>$null
if ($LASTEXITCODE -ne 0) {
  git commit -m "Initial commit"
  Write-Output "Committed initial files."
} else {
  Write-Output "Repository already has commits; skipping initial commit."
}

# Try to create remote with gh if available
if (Get-Command gh -ErrorAction SilentlyContinue) {
  if (-not $RepoName) {
    $RepoName = Split-Path -Leaf (Get-Location)
  }
  $scope = $Public.IsPresent ? "--public" : "--public"
  Write-Output "Creating GitHub repo $RepoName using gh..."
  gh repo create $RepoName $scope --source=. --remote=origin --push
  if ($LASTEXITCODE -eq 0) {
    Write-Output "Repository created and pushed."
  } else {
    Write-Warning "gh create failed. You can create the repo manually on github.com and run the push commands listed in README.md."
  }
} else {
  Write-Output "GitHub CLI (gh) not found. To create a remote, run the commands in README.md or install gh: https://cli.github.com/"
}
