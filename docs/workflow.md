## Git branching model

```txt
release-1.0          release-1.1          release-2.0
-----------          -----------          -----------
                                          * (feature-a)
                                          |
                                          * (feature-b)
                                          |
                                          | * (feature-c)
                              pick        |/
                       .----------------->* (master release-2.0)
                      /                   |
                     * fixbug-z           |
                     |        pick        |
                     |   .--------------->* (tag: v2.0.2)
        pick         |  /                 |
   .-----------------|-/----------------->* (tag: v2.0.1)
  /                  |/                   |
 /                   * fixbug-y           * (tag: v2.0.0)
'       pick         |                    |
|    .-------------->*                   /
|   /                | .----------------'
|  /                 |/  major ver incr
| /                  *
|/                   |
* fixbug-x           * (tag: v1.1.0)
|                   /
| .----------------'
|/  minor ver incr
*
|
* (tag: v1.0.0)
```

### Version

We use [semantic-version](http://semver.org/) for versioning.
In a nutshell a version is in form of: `<major>.<minor>.<patch>`,
such as `1.11.0`, `2.3.4`.

- `major` increment indicates breaking changes.
- `minor` increment indicates compatible changes.
- `patch` increment indicates bug fix etc.

These rule does not applies to versions `0.x.x`.


### Tag

Create a tag `v<major>.<minor>.<patch>` for every version and a tag should be kept for ever.


### Branch

- `release-x.y` is the branch for versions `x.y.*`.
  A `relase-x.y` branch should be kept during its maintaining period for applying bug fix to
  it or else.

- **master** is the latest **stable** branch.
  It should point to a `release-x.y` branch.

- Other branches are feature branches.
  A feature branch should only contains changes about one feature or one bug fix.


### Commit and message

**Keep a commit atomic and complete.
Every commit must pass CI**.

The `subject` of a git commit is in form of:
`<type-of-change>: <module>: what changed...`

`type-of-change` is one of:

- api-change
- new-feature
- doc
- refactor
- fixbug
- fixdoc

`module` is one of sub module or other aspect of change.
There is no strict restriction on it.

Example: `new-feature: slimtrie: add String() for print`.


## Working flow with git

1. Create a feature branch and hack on it.

   A feature branch may have more than one commits.

   Make sure every commit pass CI.

   A feature branch should base on the latest `release-x.y`.
   If a feature is introduced to a former `release-x.y`, make sure to
   `cherry-pick`(or merge) it to all newer `release-x.y`.

1. A `fix` commit(fix a bug of fix doc) should be committed to the first
   `release-x.y` branch that introduced the bug.

   And the fixbug commit should be `cherry-pick`(or `merge`) by all newer `releas-x.y`.

1. Review a feature branch by more than one members and push it to a `release-x.y` branch.

1. Bump version, create a tag for it when a `release-x.y` is ready.


### Push

- Only `fast-forward` push to `master`, `release-x.y`.

- feature branch is free to modify its history such as by `rebase` it.

- Avoid `merge`, use `rebase` instead if possible.