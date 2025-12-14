# Intelligent `drift init` Plan

## Objective
Replace the current static scaffold with an interactive, step-by-step wizard that helps the user build a tailored `.drift.yaml`. The process uses LLMs to assist with discovery and grouping, but relies on the user to validate intermediate steps to ensure quality and allows for deterministic adjustments.

## Tech Stack
*   **CLI UI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Bubbles](https://github.com/charmbracelet/bubbles).
*   **LLM:** Existing `assessor` package infrastructure.

## Core Workflow

The `init` command will launch a full-screen TUI that progresses through the following stages. **Crucially, all interactive steps will utilize static, deterministic UI elements (menus, toggles, specific input fields) rather than free-text LLM-driven adjustments for now.**

### Stage 1: Discovery (Local Analysis)
**Goal:** Identify the raw materials (files) available for the project.
1.  **Auto-Action:**
    *   Walk file system (respecting `.gitignore`).
    *   Classify files into `Docs` vs `Code` based on extensions.
2.  **Interactive Checkpoint (The "Inventory" Screen):**
    *   **Display:** Summary stats (e.g., "Found 15 documentation files and 450 code files").
    *   **User Action (Deterministic UI):**
        *   Option to view a paginated list of identified `Doc` files.
        *   Option to mark specific files as "ignore" (not a doc).
        *   Option to add a file path/glob to the `Docs` list (if missed).
        *   **Confirm/Continue:** Proceed to the next stage.

### Stage 2: Grouping (Semantic Analysis)
**Goal:** Organize the flat list of docs into logical features/rules.
1.  **Auto-Action (LLM):**
    *   Send the *names/paths* (and potentially headers) of the approved `Doc` files to the LLM.
    *   **Prompt:** "Group these documents into logical features (e.g., 'Auth', 'Billing'). Suggest a name and a doc-glob for each group. Do not merge groups that are clearly distinct."
2.  **Interactive Checkpoint (The "Features" Screen):**
    *   **Display:** A list of suggested features/rules. Each rule will show its proposed `name` and `doc` globs.
    *   **User Action (Deterministic UI):**
        *   **Rename** a feature.
        *   **Delete** a feature.
        *   **Edit Doc Globs:** Manually adjust the glob patterns for a feature.
        *   **Add New Feature:** Manually create a new feature/rule and assign doc globs.
        *   **Confirm/Continue:** Proceed to the next stage.

### Stage 3: Mapping (Code Association)
**Goal:** Find the code that implements the features defined in Stage 2.
1.  **Auto-Action (LLM):**
    *   Send the approved Feature List (from Stage 2) + the list of `Code` files to the LLM.
    *   **Prompt:** "For each Document Rule provided, identify which code file paths or directories likely implement that logic. Return specific `code` globs for each rule. Consider the file paths and directory structure as primary indicators."
2.  **Interactive Checkpoint (The "Mapping" Screen):**
    *   **Display:** The full proposed Rules (Feature Name + Doc Globs + Code Globs).
    *   **User Action (Deterministic UI):**
        *   Expand a rule to view/edit its specific `code` globs.
        *   Option to add new `code` globs to a feature.
        *   Option to remove existing `code` globs from a feature.
        *   **Confirm/Continue:** Proceed to the finalization stage.

### Stage 4: Finalize
**Goal:** Write the configuration file.
1.  **Action:** Generate the `.drift.yaml` file based on the final, user-approved state.
2.  **Display:** Success message + "Run `drift check` to start!"

## Considerations & Edge Cases

### 1. The "General" Doc Problem
*   **Strategy:** During Stage 2 (Grouping), explicitly ask the LLM to identify "General/Overview" documentation.
*   **UI:** Flag these groups in the UI. Potentially default them to *disabled* or mapped only to "entry point" files to avoid checking `README.md` against every utility function. The user can enable/disable them.

### 2. Large Repositories & Token Limits
*   **Strategy (Stage 1 & 3):** If the number of `codePaths` or `docPaths` exceeds a predefined token budget for the LLM, switch to a "Directory Summary Mode."
    *   **Mechanism:** Instead of listing every file, send a summarized directory tree to the LLM (e.g., `src/auth/ (50 files)`, `src/components/button.tsx`, `src/components/input.tsx`).
    *   **LLM Output Expectation:** The LLM should then return broader, directory-level globs (e.g., `src/auth/**`) rather than specific file globs.

## Next Steps
1.  **Scaffold TUI:** Create the basic Bubble Tea model with a "Step" state machine.
2.  **Implement Stage 1 (Discovery):** Build the file walker and the "Inventory" TUI view.
3.  **Implement Stage 2 (Grouping):** Connect the LLM for doc grouping and build the "Features" TUI view with deterministic editing capabilities.