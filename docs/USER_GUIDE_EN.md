# User guide

## Download and connect

Use the reader on **Windows x64**. No installation is required; artwork and current verification data are embedded in the EXE.

1. Download the EXE from [Releases](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/latest).
2. Close other camera applications on the PC and turn on the camera.
3. Connect the camera to the PC using a USB cable that supports data transfer.
4. Select **PC(Tether)** on the camera. For **DC-L10**, select **LUMIX Lab**.
5. Run the reader and select **CONNECT CAMERA / REFRESH**.

The reader does not change camera settings or data.
Standard supported cameras need PC(Tether); DC-L10 has been confirmed using LUMIX Lab.
`camera_profiles.json` is optional and is only needed for [separate verification data updates](CAMERA_DATA.md).

## Understanding the values

| Field | Display meaning |
|---|---|
| Counter 1 | Power / Wake Activations |
| Counter 2 | Shutter Actuations |
| Counter 3 | Flash Activations — estimated |
| Counter 4 | Power-saving Events — estimated |
| Counters 5–7 | Unidentified |

Counters 1/2 are also marked estimated outside the [verified model/firmware pairs](CAMERA_SUPPORT.md).
Estimated refers to the **meaning of the counter**, not to an invented number.

Shutter actuations can differ from the number of photographs. Power-off shutter closure can add counts on some settings.
These values do not predict remaining shutter life or failure dates.
Error descriptions use legacy service-code references; an empty error history does not certify camera health.

## Save and share

- **SAVE PNG**: choose the purpose and serial visibility, then a destination. **PUBLIC SHARE** hides the serial.
- **SAVE REPORT**: exports TXT with only the last four serial digits.
- **COPY SUMMARY**: applies the selected serial visibility and copies to the clipboard.
- After saving, use **OPEN LAST FILE / OPEN SAVE FOLDER** to locate the result.

Default folder: `%LOCALAPPDATA%\Lumerian\LUMIX Usage Info\Reports`

## Language and navigation

The initial language follows Windows: Korean for Korean display language, otherwise English.
Choose **English / 한국어 / Windows Auto** in the header. Your preference is saved.

Tab / Shift+Tab moves focus; Enter / Space activates; Esc closes a menu or goes back.
Disconnected values are marked as the last read. Reconnect and refresh to read again.

## Troubleshooting

- **Camera not found**: check the USB data cable and camera USB mode, and close other camera applications. Reconnect and select **CONNECT CAMERA / REFRESH**.
- **Saving blocked**: if Controlled Folder Access blocks a chosen folder, try the default report folder above.
- **Camera data warning**: if you placed an external JSON beside the EXE, replace it with the file from this repository or remove it to use the embedded list. Restart the reader.

## Contact

If you encounter a problem, contact us through [GitHub Issues](https://github.com/BieupHieut/LUMIX-Usage-Info/issues) or [Instagram DM @bieup_hieut](https://www.instagram.com/bieup_hieut/).
Include the camera model, firmware, reader version and symptoms or error message. You do not need to share the full serial number.
