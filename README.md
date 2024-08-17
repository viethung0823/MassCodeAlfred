# [MassCodeAlfred](https://github.com/viethung0823/MassCodeAlfred)

---

## Introduction
The official [MassCode workflow](https://github.com/massCodeIO/assistant-alfred) uses [Alfy](https://github.com/sindresorhus/alfy). The latest release, version 1.0.1, is quite large (~20 MB) just for fetching the workflow, and it even requires running npm install to function properly.

Alfy also has performance issues, so in this version, I have switched to using [awgo](https://github.com/deanishe/awgo). It is significantly faster compared to Alfy.

Feel free to open a new PR.

## Features
This workflow current support search name of snippet, folder, tag, quickview snippet, open in massCode.
1. Search name is enabled by default
2. Search folder enabled when type "f {folderName}", example: "f Default"
3. Search tag enabled when type "t {tags}", example: "t sort"
(Note: awgo only supports fuzzy search so order matter and search tag might show wrong results when that snippet has multiple tags. This only can be solved when massCode supports query by tags with apis I'll update asap when they have it) 

## Tips
In case you didn’t know using placeholder `{cursor}` in your snippet and trigger with alfred it will move your cursor at desired place.
```
console.log({cursor});
```

## Demo

![Demo](https://raw.githubusercontent.com/viethung0823/MassCodeAlfred/main/demo.gif)
