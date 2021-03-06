## lab snippet

Create a personal or project snippet

### Synopsis


Source snippets from stdin, file, or in editor from scratch
Optionally add a title & description with -m

```
lab snippet [flags]
```

### Options

```
  -g, --global                create as a personal snippet
  -h, --help                  help for snippet
  -m, --message stringArray   use the given <msg>; multiple -m are concatenated as separate paragraphs (default [-])
  -n, --name string           (optional) name snippet to add code highlighting, e.g. potato.go for GoLang
  -p, --private               make snippet private; visible only to project members (default: internal)
      --public                make snippet public; can be accessed without any authentication (default: internal)
```

### Options inherited from parent commands

```
      --no-pager   Do not pipe output into a pager
```

### SEE ALSO

* [lab](index.md)	 - lab: A GitLab Command Line Interface Utility
* [lab snippet browse](lab_snippet_browse.md)	 - View personal or project snippet in a browser
* [lab snippet create](lab_snippet_create.md)	 - Create a personal or project snippet
* [lab snippet delete](lab_snippet_delete.md)	 - Delete a project or personal snippet
* [lab snippet list](lab_snippet_list.md)	 - List personal or project snippets

###### Auto generated by spf13/cobra on 15-Mar-2021
