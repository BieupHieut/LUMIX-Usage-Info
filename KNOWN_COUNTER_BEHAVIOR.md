# Known Counter Behavior — DC-S1RM2 / Firmware 1.5

These are controlled behavior observations from a real camera, not official Panasonic counter-name definitions.

## Counter #2 — Shutter Actuations (verified behavior)
- Mechanical shutter exposure: +1
- Two mechanical actuations: +2
- Electronic shutter exposure: +0
- High Resolution test using the tested settings: +0
- Power OFF with [Power-off Shutter] = CLOSE: +1
- Sensor Cleaning + required restart: +1 total observed
- Pixel Refresh + required restart: +3 total observed with Power-off Shutter=CLOSE; exact breakdown not isolated

## Counter #1 — Power / Wake Activations (verified behavior)
- OFF → ON: +1
- Sleep → Wake cycle: +1
- USB disconnect/reconnect only: +0

## Counters #3–#7
Unidentified. Legacy candidate names may be shown only as unverified references.


# Community Validation — DC-S5 / Firmware 2.9

Contributor: **아름프로** · Tool: **v1.0.0-beta.3**

The earlier beta.3 sample established read compatibility. The project owner subsequently confirmed Shutter Actuations and Power / Wake Activations as verified for this exact model/firmware combination on 2026-09-20. Individual controlled delta logs were not supplied, so the detailed DC-S1RM2 behavior above must not be assumed to apply to DC-S5.

# Community Validation — DC-S5M2 / Firmware 3.7

Validators: **엘가, 제비동선**. Both Shutter Actuations and Power / Wake Activations were confirmed as verified by the project owner on 2026-09-20. The provided report contains raw values, not controlled delta logs; model-specific increment rules are therefore not documented.
