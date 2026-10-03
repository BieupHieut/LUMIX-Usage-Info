# LUMIX Usage Info 1.0.0 User Guide

## Connect

Close previous reader and other camera apps. Power on; connect a USB data cable to this PC. Select **PC(Tether)**, or **LUMIX Lab for DC-L10**. Run `LUMIX_Usage_Info_v1.0.0.exe` and Connect / Refresh. Models without a supported connection mode are not supported. No separate L10 connection tool is required.

## Counters

| Raw field | Display name | Meaning status |
|---|---|---|
| Counter 1 | Power / Wake Activations | Verified only for an exact listed pair; otherwise estimated |
| Counter 2 | Shutter Actuations | Verified only for an exact listed pair; otherwise estimated |
| Counter 3 | Flash Activations | Estimated; legacy STBCNT candidate |
| Counter 4 | Power-saving Events | Estimated; legacy PSVCNT candidate |
| Counters 5–7 | Unidentified | Values retained; no invented names |

Estimated refers to the **field meaning**, not a fabricated value. Modern model mappings for 3/4 are unverified. Counter 4 is not an abnormal-shutdown diagnosis. Error history is separate; an empty history does not certify camera health.

DC-S5M2 / 3.7 remains verified. Other S5M2 firmware and S9/G9M2/GH7 show estimated meanings. See [validation matrix](VALIDATION_MATRIX.md).

Shutter actuation count is not necessarily the number of photographs. Power-off shutter closure can add counts, while electronic exposures did not add counts in the S1RM2 1.5 observations. Do not generalize detailed increment rules to other models or infer remaining lifetime/failure dates.

## Save / share

PNG: choose purpose and serial visibility, then select a destination. Public share hides serials. TXT: automatically saved with unique filenames and only the last four serial digits. Clipboard summary respects the selected export visibility. Open last file / folder after a successful save.

Default folder: `%LOCALAPPDATA%\Lumerian\LUMIX Usage Info\Reports`

### Controlled Folder Access

Windows can block a program from writing to protected folders, including Documents. Version 1.0.0 defaults to Local AppData. Selecting a protected PNG destination can still be blocked. Try the default location first. If that protected destination is necessary, Windows Security → Virus & threat protection → Manage ransomware protection → Allow an app through Controlled folder access lets you select the current executable, subject to administrator policy. Do not disable protection or rely on elevation. The reader never changes security settings. [Microsoft documentation](https://learn.microsoft.com/en-us/defender-endpoint/controlled-folders).

## Language / navigation

Windows Auto uses Korean for Korean Windows display language; otherwise English. The header lets you choose Korean / English and remembers your choice. Auto follows Windows again. Tab / Shift+Tab: focus; Enter / Space: activate; Esc: close menu / Back. Scroll when the window is smaller than the content. Disconnected values are clearly marked as the last read.

## Versions / limits

App **1.0.0** tracks software features. Camera data **2026-10-02** tracks verified pairs and can be replaced separately; see [data guide](CAMERA_DATA_GUIDE.md). Community verification is not a physical hardware test of this build. Camera communication, native save dialog, Controlled Folder Access on target PCs and movement between differently scaled monitors still need real-environment confirmation.

The one-line title is `LUMERIAN-LUMIX Usage Info`. Maximize is disabled; moving, minimizing and resizing remain available. Hover repaints only the changed buttons using a complete back buffer.

## Camera list and release history

The unified camera list shows exact verified model/firmware pairs first, followed by estimated support models using the same card layout. Browse seven rows per page with the arrow buttons. “Validated by” identifies the community contributor. Verification applies only to the listed firmware; estimated support is not counter validation.

Release history shows the software version and release date: **v1.0.0 · 2026-10-03**. See CAMERA_DATA_GUIDE.md for separate camera registry maintenance.

Footer: click v1.0.0 at the left for release history, or the centered creator label for credits. Verified entries use blue; estimated entries use a neutral warm background with amber accents. Explicit status words accompany the colors. Windows display scaling applies consistently to text, artwork and input positions.

A read-only notice appears before connection and in the connected camera information card. The program does not change camera settings or data.
