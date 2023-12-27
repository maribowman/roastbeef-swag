# roastbeef-swag discord bot

![Build](https://github.com/maribowman/roastbeef-swag/actions/workflows/build.yml/badge.svg)
![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)

## organize shopping list

organizes the shopping list of the `groceries` channel. the bot can process single and multi line input.

- ### add: `<quantity> <item> <quantity>`
    - `eggs 3` or `3 eggs`
        - ```
          coffee
          bagels 4
          3 croissants
          ```

- ### delete: `(*) <id> <id> <id> <id>-<id>`
    - single delete: `5`, `2 4`
    - range delete: `3-5`, `1 3 5-8`
    - delete all (except): `*`, `* 2 4 6-9`
