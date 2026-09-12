# Book Club

Discord bot for a book club that reads one book at a time. Members register the current book and update the chapter as meetings happen.

## Language

**Server**:
The Discord server whose members read together; owns exactly one Current Book.
_Avoid_: guild, tenant, instance.

**Current Book**:
The single book a Server's club is reading, with a Name, an optional Url, and a Chapter.
_Avoid_: library, shelf, reading list.

**Chapter**:
The club's reading progress in the Current Book; a non-negative number where 0 means not started.
_Avoid_: page, progress, bookmark.

**Register**:
Setting or replacing a Server's Current Book; always starts at Chapter 0 and keeps no history.
_Avoid_: add, create.

**Update Chapter**:
Changing the Current Book's Chapter; accepts any non-negative number, forward or backward. Not found when no Current Book exists.
_Avoid_: set progress, advance.

## Relationships

- A **Server** owns exactly one **Current Book**.
- A **Current Book** has exactly one **Chapter**.
- **Register** replaces the **Current Book** and resets its **Chapter** to 0.
- **Update Chapter** requires a **Current Book**.

## Example dialogue

> **Dev:** "When a Server registers a new book, what happens to the old one?"
> **Domain expert:** "It's replaced. We read one book at a time — no history."
> **Dev:** "Can someone update the chapter before registering?"
> **Domain expert:** "No — with no Current Book there's nothing to update. That's a not-found error."
> **Dev:** "What if we're on chapter 8 but it should be 7?"
> **Domain expert:** "Update it to 7. Corrections happen, so any non-negative number is valid."

## Flagged ambiguities

- "Club" and "Server" were used loosely. Resolved: the club is the human group; **Server** is the data scope — each server, including test servers, owns its own Current Book.
