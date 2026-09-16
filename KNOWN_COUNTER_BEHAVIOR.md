# Known Counter Behavior — DC-S1RM2 / Firmware 1.5

This document records controlled observations. It is not an official Panasonic definition of the internal fields.

## Shutter Actuations

| Action | Observed change | Status / note |
|---|---:|---|
| Mechanical shutter, 1 shot | +1 | Verified |
| Mechanical shutter, 2 shots | +2 | Verified |
| Electronic shutter, 1 shot | +0 | Verified |
| High Resolution mode test | +0 | Verified in tested setup; electronic shutter path |
| Power OFF with `[Power-off Shutter] = CLOSE` | +1 | Verified physical shutter closure |
| Sensor Cleaning + required restart | +1 total observed | No separate sensor-cleaning counter identified; power-off shutter setting can contribute |
| Pixel Refresh + required restart | Additional shutter activity observed | Exact Pixel Refresh vs power-off-shutter breakdown not isolated; do not claim a fixed increment |

## Power / Wake Activations

| Action | Observed change | Status / note |
|---|---:|---|
| Camera OFF → ON | +1 | Verified |
| Sleep → wake cycle | +1 | Verified as a cycle; exact increment moment not isolated |
| USB disconnect / reconnect only | +0 | Verified |

## Raw Counter 3–7

No change was observed during the tested flash, sleep/wake, High Resolution, Sensor Cleaning, CFexpress format, clock reset, and Pixel Refresh scenarios. Their meanings remain unidentified.
