# Community Validation Results

## DC-S5 / Firmware 2.9

- **Status:** READ COMPATIBLE
- **Tool used:** LUMIX Usage Info v1.0.0-beta.3
- **Contributor:** 아름프로
- **Result:** Camera information and the two main raw counters were successfully read.
- **Observed sample:** Raw Counter 2 = 1,074 / Raw Counter 1 = 430
- **Counter semantics:** NOT behavior-verified on DC-S5.

This result confirms read compatibility only. It does **not** establish that Raw Counter 2 means Shutter Actuations or that Raw Counter 1 means Power / Wake Activations on DC-S5.

### Useful next tests
If a DC-S5 owner wishes to help, safe controlled tests include:

1. Record a baseline.
2. Take exactly one mechanical-shutter photo and read again.
3. Take exactly one electronic-shutter photo and read again.
4. Perform one sleep-to-wake cycle and read again.

Do not use service-mode write operations or unknown opcode probing.
