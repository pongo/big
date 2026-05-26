# Big

Big is a console utility for inspecting which immediate children of a folder occupy the most disk space.

## Language

**Scan Root**:
The folder whose immediate children are inspected and ranked by size.
_Avoid_: target directory, input directory

**Root Entry**:
A file or folder directly inside the **Scan Root**. Hidden files, hidden folders, symlinks, and junctions are still **Root Entries**; a folder remains one **Root Entry** even though its size includes nested contents.
_Avoid_: item, node, result

**Entry Size**:
The size used to rank a **Root Entry**. For a file it is the file size; for a folder it is the recursive total of readable contained files and folders. Symlinks and junctions do not have an **Entry Size** because Big does not follow them.
_Avoid_: weight, disk usage

**Size Ranking**:
The ordering of **Root Entries** from largest to smallest **Entry Size**. Root entries without an **Entry Size** appear after all sized entries and are ordered by name.
_Avoid_: sorting, order

**Trashed Root Entry**:
A **Root Entry** that the user moved to the operating system trash during the current Big session. It remains visible in the current **Size Ranking** but is visually distinguished from active **Root Entries**.
_Avoid_: deleted item, removed node

## Example Dialogue

Developer: "If the scan root contains `src`, `bin`, and `README.md`, how many rows should Big show?"

Domain expert: "Three root entries. `src` and `bin` include their nested contents in their entry size, but nested files are not listed as separate rows. Links are still root entries, but they appear after entries with sizes."
