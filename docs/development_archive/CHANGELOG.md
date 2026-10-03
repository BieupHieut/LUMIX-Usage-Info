# Changelog

## v1.0.0-beta.18 — 2026-10-02

Owner-confirmed L10 1.2/S1M2 1.4 shutter and power verification; combined verification page; translated Panasonic LUMIX Café Forum; no circular scrollbar dependency at content fit.

## v1.0.0-beta.17 — 2026-10-02

L10 regular reads via LUMIX Lab; exact L10 1.2/S1M2 1.4 read reports by 잠이든; counter meanings stay unverified; Special Thanks for 파나소닉 루믹스 카페 포럼 members. Removed the dedicated test entry from the regular UI. See current release notes.

## v1.0.0-beta.16 — 2026-10-02

- Add manual L10 USB A/B/C diagnostics, consent gates, strict response validation and worker cancellation.
- Add redacted JSON/TXT preview/export, including failures; synthetic sample mode remains clearly distinguishable.
- Keep S9 on the normal tether route and L10 counter meanings unverified.
- Document official Data Act clues for other models without mapping unidentified fields or claiming public service access.

## v1.0.0-beta.15 — 2026-10-01

- Reordered home actions to group refresh / PNG / TXT together.
- Added resizable window, per-monitor DPI scaling, scrolling, keyboard navigation and hover / press feedback.
- Unified enabled-state checks across pointer and keyboard activation; dim unavailable actions.
- Added session-local saved-file and folder shortcuts.
- Wrapped notices, localized PNG dialog title / filters, and exposed language preference-save errors.
- Version-specific extracted assets and smooth bitmap scaling.

## v1.0.0-beta.14 — 2026-09-30

- Fixed clipped text in the six connected-home action cards in Korean and English.
- Centered the title and description block within each card and aligned number badges vertically.
- Verified card text bounds using actual GDI font measurements.
- Centered single-line text vertically and corrected other undersized text slots, including version rows and the English validation status.
- Fixed the header character position across all pages and separated the language controls from it.
- Used the official 2D guide sticker alongside the existing 3D character states.
- Added a non-tether-camera notice and all 19 USB tether models from the supplied / official Panasonic list, retaining the three exact counter-verified combinations.

## v1.0.0-beta.13 — 2026-09-29

- Added a persistent Language menu with Windows Auto, Korean and English choices.
- Improved Korean/English typography, spacing, connection errors, verified-camera list and export layout.
- Simplified Creator & Contributors to creator, validators and feedback channels.
- Kept camera transport, counter interpretation and the three verified combinations unchanged.

## v1.0.0-beta.12 — 2026-09-28

- Displayed provisional shutter and power/wake names on other readable model/firmware combinations, with prominent unverified warnings and raw counter mapping in details and reports.
- Added automatic Korean/English UI selection from the Windows user interface language, including localized TXT reports and PNG export screens.
- Preserved the three exact verified combinations and the read-only camera transport.

## v1.0.0-beta.11 — 2026-09-28

- Refined the dark dashboard with deep red and blue accents; routine cards remain neutral.
- Added three explicit camera-to-computer USB connection steps and a model-dependent PC(Tether) menu example.
- Kept the Lumerian character in the connection and camera panels.
- Camera transport, verified counter mappings and report behavior are unchanged.

## v1.0.0-beta.10 — 2026-09-28

- Consolidated the dashboard palette into warm apricot, soft teal and neutral gray on the existing dark background.
- Reduced visual competition by making ordinary action-card outlines neutral.
- Moved the official Lumerian art into the connection and connected-camera panels, with the error state shown when disconnected.
- Updated color accents across status and supporting pages, while preserving camera reading and button behavior.

## v1.0.0-beta.9 — 2026-09-28

- Retained the dark canvas and official Lumerian mascot while adapting the supplied mobile design references for the Windows dashboard.
- Added a connection status panel, camera identity hero, distinct shutter and power/wake cards, and six numbered action cards.
- Updated hit areas to match the new card positions and kept verified/raw labels, serial privacy, page navigation and exports intact.
- Added beta.9 to the in-app version history. No camera transport or counter interpretation changed.

## v1.0.0-beta.8 — 2026-09-28

- Established **Lumerian** as the product brand and official guide character.
- Replaced all four application mascot states with the official cream-and-golden Lumerian puppy design and red hoodie.
- Added Lumerian branding to the app header, window title, credits, version history and text reports.
- Added high-resolution transparent character masters and a product brand guide.
- Kept beta.7 camera-reading, validation and privacy behavior unchanged.

## v1.0.0-beta.7 — 2026-09-20

- QA fixes: preserve the current page while a background camera check runs and completes.
- Treat unavailable serial numbers as missing: do not calculate cross-camera deltas or display `********able`.
- Removed the overlapping note below the expanded Version History list.
- Validation & Info now shows the verified connected camera and its validator when available.
- DC-S1RM2 / FW 1.5, DC-S5M2 / FW 3.7, and DC-S5 / FW 2.9 are now listed as verified for Shutter Actuations and Power / Wake Activations.
- Corrected the DC-S5M2 report from `UNVERIFIED MODEL` and raw Counter 1/2 labels to the verified status and counter names.
- Preserved the original beta.6 release as a separate package; the earlier DC-S5 read-compatible entry below records its original status.

## v1.0.0-beta.6
- Hid low-level protocol identifiers from the user-facing Technical Details header; developer documentation still retains them
- Added disconnected / last-known-data state
- Added Windows device-change background camera re-check
- Renamed entry to Validation & Info and enabled offline access
- Added single-instance protection
- Added refreshing feedback and clipboard retry
- Report serial privacy defaults to Last 4
- Added state-aware Pomeranian mascots
- Corrected beta.3 Version History entry
- Preserved read-only communication and beta.4 PNG stability architecture
- Added first external community result: DC-S5 / FW 2.9 = READ COMPATIBLE (아름프로, tested with beta.3); raw counter semantics remain unverified
