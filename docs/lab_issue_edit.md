## lab issue edit

Edit or update an issue

```
lab issue edit [remote] <id>[:<comment_id>] [flags]
```

### Examples

```
lab issue edit <id>                                # update issue via $EDITOR
lab issue update <id>                              # same as above
lab issue edit <id> -m "new title"                 # update title
lab issue edit <id> -m "new title" -m "new desc"   # update title & description
lab issue edit <id> -l newlabel --unlabel oldlabel # relabel issue
lab issue edit <id>:<comment_id>                   # update a comment on MR
```

### Options

```
  -m, --message stringArray   use the given <msg>; multiple -m are concatenated as separate paragraphs
  -l, --label strings         add the given label(s) to the issue
      --unlabel strings       remove the given label(s) from the issue
  -a, --assign strings        add an assignee by username
      --unassign strings      remove an assignee by username
      --milestone string      set milestone
      --force-linebreak       append 2 spaces to the end of each line to force markdown linebreaks
  -h, --help                  help for edit
```

### Options inherited from parent commands

```
      --no-pager   Do not pipe output into a pager
```

### SEE ALSO

* [lab issue](lab_issue.md)	 - Describe, list, and create issues

###### Auto generated by spf13/cobra on 15-Mar-2021
