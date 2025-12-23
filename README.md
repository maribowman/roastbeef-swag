# Roastbeef-Swag Discord Bot

![Build](https://github.com/maribowman/roastbeef-swag/actions/workflows/build.yml/badge.svg)
![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)

## Organize my grocery list and freezer inventory

Track the inventory of my `groceries` and `freezer` Discord channels.

- ### Add Items

  - Use `<quantity> <item>` or `<item> <quantity>` to add an item
    - Single line: `eggs 3` or `3 eggs`
    - Multi line:

      ```
      coffee
      bagels 4
      3 croissants
      ```

  - Suggest an item via `? <quantity> <item>`. The item remains untracked in a separate message and can be added by reacting with 👍 or 👎.

- ### Remove Items

  - All items \[except\] `* [<id> ...]`
    - E.g. `*` or `* 2 4`
  - Single items `<id> <id> ...`
  - Ranges `<id>-<id>` or `<id> <id>-<id>`

### Edit Items

- Use `📝 Edit` button to edit the entire inventory list via Modal
- Edit quantity of single items
  - Increase `4++` or lower `3--` quantity
