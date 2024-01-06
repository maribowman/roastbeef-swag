# Roastbeef-Swag Discord Bot

![Build](https://github.com/maribowman/roastbeef-swag/actions/workflows/build.yml/badge.svg)
![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)

## Organize grocery list and frozen inventory

Organizes the shopping list my `groceries` channel and my `tk-goods` frozen inventory. The bot can process single and
multi line input.

- ### Add: `<quantity> <item> <quantity>`
    - `eggs 3` or `3 eggs`
    - ```
      coffee
      bagels 4
      3 croissants
       ```

- ### Remove: `(*) <id> <id> <id> <id>-<id>`
    - Single: `5`, `2 4`
    - Range: `3-5`, `1 3 5-8`
    - All (except): `*`, `* 2 4 6-9`
