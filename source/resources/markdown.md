# Markdown

Select across **paragraphs**, *styles*, ~~corrections~~ and `code`.
中文也可以选择。日本語と emoji 🌿 also remain selectable.

```counter title=Interactive
Application-owned counter
```

[Example action](example:action)

## Image

![Author avatar](images/avatar.jpeg "Local gallery resource")

## Text and links

A paragraph with **bold and *nested emphasis***, escaped \*asterisks\*,
&amp; entities and `code with <angle brackets>`.
A backslash creates a hard break here.\
This starts a new line in the same paragraph.

[Named action][action] and <https://example.com> exercise link activation.

Inline math such as \(a^2 + b^2 = c^2\) and \(2\pi \cdot r\) is drawn by the application.

[action]: example:action "Gallery action"

## Lists and quotes

- First item
  - A nested item with **emphasis**
  - Another nested item
- Last item

3. Ordered lists can start at three.
4. Numbering continues.

- [x] Select text across blocks
- [ ] Try copying a long selection

> A quoted paragraph with **formatted text**.
>
> > A nested quote that wraps on narrow windows.

## Table

| Content | Alignment | Value |
| :--- | :---: | ---: |
| Text | Center | 42 |
| **Bold** and `code` | Center | 1,024 |
| 中文 | 居中 | 128 |

## Code

```go
func main() {
    fmt.Println("Hello, Onevra")
}
```

```typescript
interface User { name: string; active: boolean }
const welcome = (user: User) => `Hello, ${user.name}`;
```

```python
# Unicode and multiline strings retain their source text.
def greet(name):
    return f"Hello, {name} 世界"
```

```json
{ "name": "Onevra", "readonly": true, "languages": 18 }
```

```unknown
An unregistered language remains a selectable code block.
```

    Indented code is also supported.
    Its whitespace is preserved.

## Image fallbacks

![Missing image — alternative text](images/missing.png)

An inline image ![Inline avatar](images/avatar.jpeg) uses its alternative text
so the surrounding sentence remains a single selectable paragraph.

## Literal HTML

<div>Raw HTML stays visible as text.</div>

---

End of the document. Select All includes content outside the viewport.
