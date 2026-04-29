# Wiki Writing Guide

`docs/wiki/` pages should help readers find the right command, configuration, or troubleshooting step quickly.

## Core rules

- Put the answer or next action in the first paragraph.
- Start body sections at `##`; the page title is the only level-1 heading.
- Use descriptive headings such as `## Configure JWT authentication`, not `## Details`.
- Keep paragraphs short and focused on one idea.
- Use numbered lists for steps, bullets for options, and tables only when they make comparisons easier.
- Use descriptive link text that makes sense outside the surrounding sentence.
- Do not use bold text, extra blank lines, or manual indentation as substitutes for real headings, lists, or tables.

## Page structure

Use this default order unless the page has a clear reason to differ:

1. One-sentence purpose statement
2. Prerequisites or when to use the page
3. Main steps or reference table
4. Examples
5. Troubleshooting or related pages

## Headings

Use heading levels in sequence. After the page title, start with `##` sections, then use `###` only for subsections inside the current `##` section.

Name headings after the reader's task or destination. Prefer `## Configure JWT authentication` or `## Find pagination fields` over vague names such as `## Details`, `## Notes`, or `## More`.

## Links

Use link text that describes the destination or action.

| Avoid | Prefer |
|---|---|
| `[here](Authentication-and-Profiles.md)` | `[Authentication and Profiles](Authentication-and-Profiles.md)` |
| `[details](Output-and-Pagination.md)` | `[Output and Pagination](Output-and-Pagination.md)` |

## Lists and tables

Use numbered lists when the reader must complete steps in order. Use bullets for options, alternatives, or related notes that do not require a sequence.

Use tables only when clear headers reduce comparison effort. If a table does not make values easier to compare, use a short list instead.

## Examples

Place examples near the rule or command they explain. Keep examples short enough that readers can copy the command and understand the expected result without scanning unrelated context.

Pattern:

1. State the command purpose.
2. Show the command.
3. Explain the important output or next action.

## What to avoid

- Fake headings made from bold text, extra blank lines, or punctuation.
- Vague links such as `here`, `details`, or `this page`.
- Manual indentation used to imitate lists, tables, or code blocks.
- Long paragraphs that combine setup, commands, output, and troubleshooting in one block.
