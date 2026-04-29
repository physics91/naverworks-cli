# Wiki Review Checklist

Use this checklist before publishing changed pages from `docs/wiki/` to the GitHub wiki.

## Reader task

- [ ] The title clearly describes the page purpose.
- [ ] The first paragraph says who should use the page and what they can do next.
- [ ] A new reader can find installation, authentication, command usage, or troubleshooting paths without reading the whole page.

## Structure

- [ ] The page has exactly one level-1 heading.
- [ ] Body sections start at `##`.
- [ ] Heading levels do not skip, such as `##` to `####`.
- [ ] No bold text is used as a fake heading.

## Scannability

- [ ] Long paragraphs are split into shorter paragraphs or lists.
- [ ] Steps use numbered lists.
- [ ] Options, commands, or fields use bullets or tables.
- [ ] The most important information appears before background details.

## Links

- [ ] Link text describes the destination or action.
- [ ] The page does not use vague links like `here`, `details`, or `more`.
- [ ] Related pages are grouped where they help the reader decide the next step.

## Accessibility

- [ ] Lists use Markdown list syntax, not manual indentation.
- [ ] Tables have clear headers.
- [ ] Images include meaningful alt text; decorative images are omitted.

## Maintenance

- [ ] Version-specific or time-sensitive statements include the affected version, date, or condition.
- [ ] Troubleshooting content appears in a clearly named section or related-page link.
- [ ] New pages are added to `_Sidebar.md` when they should be discoverable.
