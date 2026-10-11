# LUMERIAN LUMIX Usage Board — v1.1.1

Windows x64 · Build · 2026-10-12. Download and run the EXE; no installation is required. Artwork and camera verification data are embedded. The optional camera_profiles.json updates the verification list separately.

### Navigation

The same four tabs stay at the bottom, above the language footer before and after connection: **Camera / Usage history / Connection & diagnostics / Compatibility**.

- Camera: connection and counts; Save report, Technical details and Camera error history.
- Usage history: records, camera filter, cleanup and Save settings.
- Connection & diagnostics: retry, received values and diagnostic ZIP.
- Compatibility: verified and estimated camera combinations.

App information stays at the bottom right. Click the version beside the product name for version history. **The round left-arrow button returns to the screen you entered from** (Escape or Alt+Left also works), including nested pages. Choosing a main tab starts a new navigation path. The Compatibility tab has no duplicate App information or Back buttons.

### Connect and read

1. Close LUMIX Tether, other camera applications and older copies of this tool.
2. Turn on the camera, connect it by USB and select **PC(Tether)**. For DC-L10, select **LUMIX Lab**.
3. On the Lumerian welcome screen, click **Connect camera**.

The tool reads information without changing camera settings or saved data, and releases the device after each query. Check Camera compatibility for supported and estimated models. DC-S1 / firmware 2.4 counter support remains on hold.

### Three information cards

- **My camera:** model, firmware and serial visibility.
- **Shutter actuations:** cumulative shutter count.
- **Power / wake activations:** cumulative power/wake count, not battery information.

Unverified meanings retain an Estimated / unverified label; unavailable counts appear as a dash. Counts do not predict remaining shutter life. Click **Read again** to refresh.

Click the serial box to cycle **Hide all → Last four digits → Show all**. Each new query hides it again. Export privacy is selected separately.

### History and sharing

Successful counter reads are saved locally with a timestamp. **Usage history** works without a connected camera. Choose a body to view its records and changes; multiple bodies of the same model remain separate. History retains the serial suffix and a local identifier, not the full serial. Without a reliable serial, automatic history is not saved.

In **Usage history → Save settings**, export all history to JSON or update the selected copy. Automatic updates follow successful history saves. External file changes stop automatic overwriting until the destination is selected again. Internal history stays in Local AppData. Displayed dates use `YYYY-MM-DD HH:MM:SS`; filenames use `YYYYMMDD_HHMMSS`.

**Save report** provides PNG, TXT and Copy summary, all using the chosen serial privacy. Each time you open this screen, the serial starts fully hidden; choose Last four or Show all again if needed. PNG prompts for a destination. TXT goes to the local report folder; Open file / Open folder is available afterwards.

Click **Clean duplicates** to remove intermediate records whose shutter AND power counts match both neighbouring observations on the same local date, for the selected body (or all cameras). The first/latest observations, daily endpoints and both sides of counter changes are retained, as are firmware, verification and reset boundaries. Review the deletion count before confirming. A complete local backup is created in `history-backups`; **History folder** takes you there. A failed backup cancels cleanup. External files are kept; their automatic updates are disabled after cleanup, until you explicitly enable them again. To restore a backup, close the app, preserve the current `usage-board.json`, and copy the chosen backup over it.

Hover over or click **!** beside Shutter actuations for the shutdown-shutter explanation. Click again or press Escape to close. The official setting is [Shutter Behavior at Power Off](https://eww.pavc.panasonic.co.jp/dscoi/DC-S1RM2/html/DC-S1RM2_DVQP3245_eng/0130.html). The +1 observation comes from this project's DC-S1RM2 / firmware 1.5 test, not an official counter guarantee for all cameras.

### Select and delete records

Check individual rows in **Usage history**, then click **Delete selected (n)**. Selections remain when paging; changing the camera filter clears them. **Clear selection** unchecks all selected rows. Review the count and confirm before deleting. A complete backup is saved in `history-backups` first; failure to back up cancels deletion. External history copies remain unchanged and automatic overwriting is switched off.

Deletion also marks the removed interval so it is not joined into a misleading increase. After the first selected deletion, manage this history with **v1.1.0 or newer**; earlier builds cannot read its updated format. The backup made before deletion retains the original format. Use the restoration steps above if needed.

### Daily, monthly and yearly analysis

Open **Usage history → Analysis**, choose one camera body, then **Daily / Monthly / Yearly** and **Shutter / Power**. Eight periods are shown at a time; use Previous / Next for older or newer periods. Analysis is available without a connected camera.

Bars show observed counter increases between comparable readings within the same period, not a complete count of photos actually taken in that period. **0** means comparable readings confirmed no increase; **—** means there is no comparable interval. Keep readings over time to build useful comparisons. PC local dates determine the period boundaries.

An increase between readings in different periods is shown separately below the chart, across the selected body's entire history, and is not assigned to a bar. It can become comparable in a broader monthly or yearly period. Firmware changes, counter resets, verification changes and deleted intervals are excluded from comparisons. Estimated readings retain an estimated label. Selecting Shutter or Power changes the chart scale independently.

### Connection problems

If the app reports a startup failure before its main window opens, include the displayed stage and code when contacting support. If it is already running, use the open window.

Open **Connection & diagnostics** using its fixed bottom tab. Select a camera if multiple devices are detected. A running query can be stopped here.

**Received values** shows connection and query stage results. **Save diagnostic ZIP** saves the latest query, so save a particular failure before retrying. The ZIP excludes full serials, device paths, card labels, filenames and complete raw responses. Nothing is uploaded automatically.

Send the ZIP, symptoms, model and firmware through [GitHub Issues](https://github.com/BieupHieut/LUMIX-Usage-Info/issues) or Instagram DM **@bieup_hieut**.

### Language

The footer offers Windows Auto, Korean, English, Japanese and Spanish. Other Windows languages default to English. Your choice persists.
