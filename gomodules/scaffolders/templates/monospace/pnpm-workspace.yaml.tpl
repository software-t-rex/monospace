packages:
  # all packages in direct subdirs of apps/
  - 'apps/*'
  # all packages in direct subdirs of packages/
  - 'packages/*'
  # exclude packages that are inside test directories
  - '!**/test/**'