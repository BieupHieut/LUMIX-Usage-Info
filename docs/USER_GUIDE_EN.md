# User guide

## Download and connect

1. Download the EXE from [Releases](https://github.com/BieupHieut/LUMIX-Usage-Info/releases/latest). No installation is required.
2. Close other camera applications and power on the camera.
3. Connect it to the PC using a USB data cable.
4. Select **PC(Tether)** on the camera, or **LUMIX Lab for DC-L10**.
5. Run the reader and select **Connect / Refresh**.

Standard supported cameras need PC(Tether). DC-L10 has been confirmed using LUMIX Lab.
The reader does not change camera settings or data.

## Understanding the values

| Field | Display meaning |
|---|---|
| Counter 1 | Power / Wake Activations |
| Counter 2 | Shutter Actuations |
| Counter 3 | Flash Activations — estimated |
| Counter 4 | Power-saving Events — estimated |
| Counters 5–7 | Unidentified |

Counters 1/2 are also marked estimated outside the [verified model/firmware pairs](CAMERA_SUPPORT.md).
Estimated refers to the **meaning of a counter**, not to an invented number.

Shutter actuation count can differ from the number of photographs. Power-off shutter closure can add counts on some settings.
These values do not predict remaining shutter life or failure dates.
Error descriptions use legacy service-code references; an empty history does not certify camera health.

## Save and share

- **Save PNG**: choose the purpose and serial visibility, then a destination. Public sharing hides the serial.
- **Save report**: exports TXT with only the last four serial digits.
- **Copy summary**: applies the selected visibility and copies to the clipboard.
- Use **Open file / Open folder** after saving to locate the result.

Default folder: `%LOCALAPPDATA%\Lumerian\LUMIX Usage Info\Reports`

## Language and navigation

The initial language follows Windows: Korean for Korean display language, otherwise English.
Choose **Korean / English / Windows Auto** in the header. Your preference is saved.

Tab / Shift+Tab moves focus; Enter / Space activates; Esc closes menus or goes back.
Disconnected values are marked as the last read. Reconnect and refresh to read again.

## Troubleshooting

- Check the USB data cable and camera USB mode, and close other camera applications.
- If Controlled Folder Access blocks saving, try the default report folder first. The reader does not change Windows security settings.
- Report persistent problems through [GitHub Issues](https://github.com/BieupHieut/LUMIX-Usage-Info/issues) with the model, firmware, reader version and symptoms. Do not publish the full serial number.

## Camera data updates

Current verification data is embedded in the EXE. `camera_profiles.json` is optional.
To apply additional verified pairs, place the updated JSON from this repository beside the EXE and restart.
[Details](CAMERA_DATA.md)
