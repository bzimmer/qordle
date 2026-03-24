'use strict';

/* ==============================================
   Constants & State
   ============================================== */
const WORD_LENGTH = 5;

const STATES = {
  EMPTY:   'empty',
  ABSENT:  'absent',   // gray  — letter not in word
  PRESENT: 'present',  // yellow — letter in wrong position
  CORRECT: 'correct',  // green  — letter in correct position
};

// Clicking a filled tile cycles through these states
const STATE_CYCLE = {
  [STATES.ABSENT]:  STATES.PRESENT,
  [STATES.PRESENT]: STATES.CORRECT,
  [STATES.CORRECT]: STATES.ABSENT,
};

// Counter used to give each row a unique numeric ID
let rowCounter = 0;

/* ==============================================
   Tile Helpers
   ============================================== */

/**
 * Set the letter displayed in a tile.
 * Clears the tile when `letter` is an empty string.
 */
function setTileLetter(tile, letter) {
  tile.dataset.letter = letter;
  tile.textContent = letter ? letter.toUpperCase() : '';

  if (!letter) {
    tile.dataset.state = STATES.EMPTY;
    tile.className = 'tile';
  } else {
    if (tile.dataset.state === STATES.EMPTY) {
      tile.dataset.state = STATES.ABSENT;
    }
    tile.classList.remove('tile--pop');
    // Trigger reflow so the animation restarts on each keystroke
    void tile.offsetWidth;
    tile.classList.add('tile--filled', `tile--${tile.dataset.state}`, 'tile--pop');
  }

  tile.setAttribute('aria-label',
    letter
      ? `${letter.toUpperCase()}, ${tile.dataset.state}, position ${Number(tile.dataset.col) + 1}`
      : `Empty, position ${Number(tile.dataset.col) + 1}`
  );
}

/**
 * Update only the colour-state of an already-filled tile.
 */
function setTileState(tile, state) {
  const prev = tile.dataset.state;
  if (prev !== STATES.EMPTY) {
    tile.classList.remove(`tile--${prev}`);
  }
  tile.dataset.state = state;
  if (state !== STATES.EMPTY) {
    tile.classList.add(`tile--${state}`);
  }
  if (tile.dataset.letter) {
    tile.setAttribute('aria-label',
      `${tile.dataset.letter.toUpperCase()}, ${state}, position ${Number(tile.dataset.col) + 1}`
    );
  }
}

/* ==============================================
   Row Management
   ============================================== */

/**
 * Move keyboard focus to the tile at `col` within `row`, updating
 * the roving-tabindex so Tab navigation feels natural.
 */
function focusTileInRow(row, col) {
  const tiles = row.querySelectorAll('.tile');
  tiles.forEach((t, i) => t.setAttribute('tabindex', i === col ? '0' : '-1'));
  tiles[col].focus();
}

/** Handle a keydown event on a single tile. */
function onTileKeydown(e) {
  const tile = e.currentTarget;
  const col  = Number(tile.dataset.col);
  const row  = tile.closest('.guess-row');

  switch (e.key) {
    case 'Backspace':
    case 'Delete': {
      e.preventDefault();
      if (tile.dataset.letter) {
        setTileLetter(tile, '');
      } else if (col > 0) {
        const prev = row.querySelector(`[data-col="${col - 1}"]`);
        setTileLetter(prev, '');
        focusTileInRow(row, col - 1);
      }
      break;
    }
    case 'ArrowLeft':
      e.preventDefault();
      if (col > 0) focusTileInRow(row, col - 1);
      break;
    case 'ArrowRight':
      e.preventDefault();
      if (col < WORD_LENGTH - 1) focusTileInRow(row, col + 1);
      break;
    case 'Enter':
      e.preventDefault();
      fetchSuggestions();
      break;
    default:
      if (/^[a-zA-Z]$/.test(e.key)) {
        e.preventDefault();
        setTileLetter(tile, e.key.toLowerCase());
        // Auto-advance focus to the next empty tile, or stay at last
        if (col < WORD_LENGTH - 1) focusTileInRow(row, col + 1);
      }
      break;
  }
}

/** Handle a click on a tile — cycle state if filled, otherwise focus. */
function onTileClick(e) {
  const tile = e.currentTarget;
  // Make sure tab stops at this tile
  const row = tile.closest('.guess-row');
  focusTileInRow(row, Number(tile.dataset.col));

  if (tile.dataset.letter) {
    const next = STATE_CYCLE[tile.dataset.state] || STATES.ABSENT;
    setTileState(tile, next);
  }
}

/** Update visibility of all remove buttons (hidden when only 1 row). */
function syncRemoveButtons() {
  const rows = document.querySelectorAll('.guess-row');
  rows.forEach(r => {
    const btn = r.querySelector('.guess-row__remove');
    if (btn) btn.style.visibility = rows.length > 1 ? 'visible' : 'hidden';
  });
}

/**
 * Create a new guess row, optionally pre-filled with `word` (all absent).
 * Returns the row element.
 */
function addGuessRow(word) {
  const rowId   = rowCounter++;
  const row     = document.createElement('div');
  row.className = 'guess-row';
  row.dataset.row = rowId;

  const tilesDiv = document.createElement('div');
  tilesDiv.className = 'guess-row__tiles';
  tilesDiv.setAttribute('role', 'group');
  tilesDiv.setAttribute('aria-label', `Guess ${document.querySelectorAll('.guess-row').length + 1}`);

  for (let col = 0; col < WORD_LENGTH; col++) {
    const tile = document.createElement('button');
    tile.type = 'button';
    tile.className = 'tile';
    tile.dataset.row   = rowId;
    tile.dataset.col   = col;
    tile.dataset.state = STATES.EMPTY;
    tile.dataset.letter = '';
    tile.setAttribute('aria-label', `Empty, position ${col + 1}`);
    tile.setAttribute('tabindex', col === 0 ? '0' : '-1');
    tile.addEventListener('click', onTileClick);
    tile.addEventListener('keydown', onTileKeydown);
    tilesDiv.appendChild(tile);
  }

  const removeBtn = document.createElement('button');
  removeBtn.type = 'button';
  removeBtn.className = 'guess-row__remove';
  removeBtn.setAttribute('aria-label', 'Remove this guess');
  removeBtn.textContent = '×';
  removeBtn.addEventListener('click', () => {
    row.remove();
    syncRemoveButtons();
  });

  row.appendChild(tilesDiv);
  row.appendChild(removeBtn);
  document.getElementById('guessRows').appendChild(row);
  syncRemoveButtons();

  // Pre-fill with a word if provided
  if (word) {
    const { correctByPosition, presentLetters } = buildKnowledge();
    const tiles = tilesDiv.querySelectorAll('.tile');
    [...word.toLowerCase().slice(0, WORD_LENGTH)].forEach((ch, i) => {
      if (i < WORD_LENGTH) {
        setTileLetter(tiles[i], ch);
        // Auto-apply known state from prior rows
        if (correctByPosition[i] === ch) {
          setTileState(tiles[i], STATES.CORRECT);
        } else if (presentLetters.has(ch)) {
          setTileState(tiles[i], STATES.PRESENT);
        }
      }
    });
    // Focus first tile so user can immediately adjust colors if needed
    focusTileInRow(row, 0);
  } else {
    tilesDiv.querySelector('.tile').focus();
  }

  return row;
}

/* ==============================================
   Knowledge Inference
   ============================================== */

/**
 * Scan all existing completed rows and return:
 *   correctByPosition  — Map<col, letter> for every CORRECT tile seen
 *   presentLetters     — Set of letters marked PRESENT in any prior row
 *
 * Used to auto-color tiles when a suggestion is inserted as a new row.
 */
function buildKnowledge() {
  const correctByPosition = {};
  const presentLetters    = new Set();

  document.querySelectorAll('.guess-row').forEach(row => {
    row.querySelectorAll('.tile').forEach(tile => {
      if (!tile.dataset.letter) return;
      const col    = Number(tile.dataset.col);
      const letter = tile.dataset.letter;
      if (tile.dataset.state === STATES.CORRECT) {
        correctByPosition[col] = letter;
      } else if (tile.dataset.state === STATES.PRESENT) {
        presentLetters.add(letter);
      }
    });
  });

  return { correctByPosition, presentLetters };
}

/* ==============================================
   Encode a Row → API Notation
   ============================================== */

/**
 * Encode one guess row into qordle notation:
 *   CORRECT  → uppercase letter      (e.g. "A")
 *   PRESENT  → "." + lowercase       (e.g. ".a")
 *   ABSENT   → lowercase             (e.g. "a")
 *
 * Returns null if any tile is empty (row is incomplete).
 */
function encodeRow(row) {
  const tiles = [...row.querySelectorAll('.tile')];
  if (tiles.some(t => !t.dataset.letter)) return null;
  return tiles.map(t => {
    const letter = t.dataset.letter;
    switch (t.dataset.state) {
      case STATES.CORRECT: return letter.toUpperCase();
      case STATES.PRESENT: return '.' + letter.toLowerCase();
      default:             return letter.toLowerCase(); // ABSENT
    }
  }).join('');
}

/* ==============================================
   API & Results
   ============================================== */

async function fetchSuggestions() {
  const rows    = [...document.querySelectorAll('.guess-row')];
  const guesses = rows.map(encodeRow).filter(Boolean);

  const loadingEl = document.getElementById('loadingState');
  const errorEl   = document.getElementById('errorState');
  const resultsEl = document.getElementById('resultsArea');
  const suggestBtn = document.getElementById('suggestBtn');

  loadingEl.hidden = false;
  errorEl.hidden   = true;
  resultsEl.hidden = true;
  suggestBtn.disabled = true;

  try {
    // Build URL: each guess is individually encoded; guesses are joined with %20 (encoded space)
    const path = guesses.map(g => encodeURIComponent(g)).join('%20');
    const url  = path ? `/qordle/suggest/${path}` : '/qordle/suggest/';
    const res  = await fetch(url, { method: 'POST' });

    if (!res.ok) throw new Error(`Server returned ${res.status}`);

    const suggestions = await res.json();
    displaySuggestions(suggestions);
  } catch (err) {
    errorEl.textContent = `Could not fetch suggestions: ${err.message}`;
    errorEl.hidden = false;
  } finally {
    loadingEl.hidden = true;
    suggestBtn.disabled = false;
  }
}

function displaySuggestions(suggestions) {
  const grid     = document.getElementById('suggestionGrid');
  const countEl  = document.getElementById('resultsCount');
  const resultsEl = document.getElementById('resultsArea');

  grid.innerHTML = '';

  if (!suggestions || suggestions.length === 0) {
    grid.innerHTML = '<p class="no-suggestions">No suggestions found — try adjusting your guess colours.</p>';
    countEl.textContent = '';
  } else {
    countEl.textContent = suggestions.length;
    suggestions.forEach(word => {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'suggestion-tile';
      btn.textContent = word;
      btn.setAttribute('aria-label', `Use "${word}" as next guess`);
      btn.addEventListener('click', () => {
        addGuessRow(word);
        // Scroll the new row into view
        const newRow = document.querySelector('.guess-row:last-child');
        newRow?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        // Brief visual feedback on the tile
        btn.classList.add('suggestion-tile--copied');
        setTimeout(() => btn.classList.remove('suggestion-tile--copied'), 800);
      });
      grid.appendChild(btn);
    });
  }

  resultsEl.hidden = false;
}

/* ==============================================
   Clear All
   ============================================== */

function clearAll() {
  rowCounter = 0;
  document.getElementById('guessRows').innerHTML = '';
  document.getElementById('loadingState').hidden = true;
  document.getElementById('errorState').hidden   = true;
  document.getElementById('resultsArea').hidden  = true;
  addGuessRow();
}

/* ==============================================
   Initialise
   ============================================== */

document.addEventListener('DOMContentLoaded', () => {
  // Seed the first guess row
  addGuessRow();

  document.getElementById('addGuessBtn').addEventListener('click', () => addGuessRow());
  document.getElementById('clearBtn').addEventListener('click', clearAll);
  document.getElementById('suggestBtn').addEventListener('click', fetchSuggestions);

  // Collapsible help section
  const helpToggle  = document.getElementById('helpToggle');
  const helpContent = document.getElementById('helpContent');
  helpToggle.addEventListener('click', () => {
    const expanded = helpToggle.getAttribute('aria-expanded') === 'true';
    helpToggle.setAttribute('aria-expanded', String(!expanded));
    helpContent.hidden = expanded;
  });
});
