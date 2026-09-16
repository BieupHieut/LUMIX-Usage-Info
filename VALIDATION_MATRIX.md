# Camera Validation Matrix

Validation levels are intentionally separated:

- **FULLY VERIFIED** — camera read succeeds and interpreted counter meanings were behavior-tested on real hardware.
- **READ COMPATIBLE** — usage data was successfully read from real hardware, but raw counter meanings have not yet been behavior-tested.
- **NEEDS VALIDATION** — no real-camera result has been recorded by this project yet.

## Fully verified
| Model | Firmware | Tool | Status | Contributor |
|---|---|---|---|---|
| DC-S1RM2 | 1.5 | v1.0.0-beta.6 + prior controlled tests | **FULLY VERIFIED** — read path and Counter #1/#2 behavior tested | @bieup_hieut |

## Read compatible
| Model | Firmware | Tool used for confirmation | Status | Contributor |
|---|---|---|---|---|
| DC-S5 | 2.9 | v1.0.0-beta.3 | **READ COMPATIBLE** — Setup/usage data read succeeded; counter meanings unverified | 아름프로 |

For DC-S5, keep the UI labels as **Raw Counter 2** and **Raw Counter 1** until controlled behavior tests establish their meanings. Do not infer SHTCNT/PWRCNT semantics solely because the values are returned.

## Needs hardware validation
These are candidates for testing only. Their USB transport and counter field meanings are **not** assumed compatible.

- DC-S1M2
- DC-S1M2ES
- DC-S5M2X
- DC-S5M2
- DC-S9
- DC-GH7
- DC-G9M2
- DC-L10

Do not label a model fully verified merely because the read command returns data. Confirm each interpreted field with controlled, read-only tests.
