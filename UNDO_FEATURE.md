# Undo Feature Documentation

## Overview

The 2n tracker now includes comprehensive undo functionality that captures and restores previous states for all major data modifications.

## Usage

- **Ctrl+Z** - Undo the last modification

## What Gets Captured

The undo system automatically captures state before these operations:

### Data Modifications
- **Cell value changes** (Ctrl+Up/Down/Left/Right)
- **Cell clearing** (Backspace) 
- **Row deletion** (Ctrl+H)
- **Song chain assignments**
- **Chain phrase assignments**
- **Phrase note/parameter editing**

### Paste Operations
- **Cell paste** (Ctrl+V)
- **Row paste** (Ctrl+V with row data)
- **Last edited row paste** (S key)

## What Gets Restored

Each undo operation restores:

1. **Data State** - The exact values that were modified
2. **Cursor Position** - Row, column, phrase, chain, track
3. **View Context** - Current view mode and scroll position

## Technical Details

### Memory Optimization
- Only captures the specific data type being modified (song/chain/phrase)
- Avoids storing the entire application state for each undo

### History Management
- Maintains up to 50 undo states by default
- Automatically removes oldest entries when limit is exceeded
- History is cleared on application restart

### Data Types Tracked
- **Song data**: Chain assignments per track/row
- **Chain data**: Phrase assignments per chain/row (separate for Instrument/Sampler)
- **Phrase data**: Note and parameter data per phrase/row (separate for Instrument/Sampler)
- **Associated files**: Sampler phrase file associations

## Implementation Notes

The undo system is integrated at the lowest level of data modification functions, ensuring comprehensive coverage without requiring changes to higher-level UI code. State capture happens automatically before any destructive operation, making the feature transparent to normal workflow.