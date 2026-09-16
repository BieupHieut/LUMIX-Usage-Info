# Read-only Safety Notes

LUMIX Usage Info is intentionally a read-only utility.

Allowed camera operations in the public tool:

- PTP `0x1001` GetDeviceInfo
- Panasonic read operation `0x9414` with tag `0x15C00010`

Do not add the following to a public build:

- `0x9403` SetProperty
- `0x940B`
- `0x9704` SetMntInfo
- EEPROM / ROM writes
- Firmware modification
- Unverified service-mode write actions

Unknown opcodes should not be brute-forced from this application.
