# 기종 지원과 검증 / Camera support

## 검증된 조합 / Verified pairs

| 모델 / Model | 펌웨어 / Firmware | 셔터 / Shutter | 전원·깨우기 / Power-wake | 확인 / Confirmed by |
|---|---|---|---|---|
| DC-S1RM2 | 1.5 | ✓ | ✓ | @bieup_hieut |
| DC-S5M2 | 3.7 | ✓ | ✓ | 엘가, 제비동선 |
| DC-S5 | 2.9 | ✓ | ✓ | 아름프로 |
| DC-L10 | 1.2 | ✓ | ✓ | 잠이든 |
| DC-S1M2 | 1.4 | ✓ | ✓ | 잠이든 |

위 조합에서 Counter 1은 전원·깨우기, Counter 2는 셔터 작동 횟수로 확인됐습니다.
커뮤니티 확인 결과이며, **같은 모델의 다른 펌웨어까지 검증된 것은 아닙니다**.
Counter 3/4는 추정, 5–7은 의미 미확인입니다.

Counters 1/2 are confirmed as power/wake and shutter actuations for these community-verified pairs.
**Other firmware on the same model is not automatically verified.** Counters 3/4 remain estimated and 5–7 unidentified.

## 지원 추정 목록 / Estimated support

PC(테더) 연결을 바탕으로 한 지원 추정 목록입니다. 모든 기종의 조회 성공이나 카운터 의미를 보증하지 않습니다.
아래 모델에서 확인된 결과가 있으면 [기종 확인 제보](https://github.com/BieupHieut/LUMIX-Usage-Info/issues/new/choose)를 부탁드립니다.

These are PC(Tether)-based support candidates, not confirmed reads or counter meanings for every model.
Please submit results through [Issues](https://github.com/BieupHieut/LUMIX-Usage-Info/issues/new/choose).

| 시리즈 / Series | 미검증 모델 / Unverified models |
|---|---|
| S | DC-BS1H · DC-S1 · DC-S1R · DC-S1M2ES · DC-S1H · DC-S5M2X · DC-S9 |
| G / GH / BGH | DC-BGH1 · DC-GH5 · DC-GH5S · DC-GH5M2 · DC-G9 · DC-G9M2 · DC-GH6 · DC-GH7 |

## 연결 모드 / USB mode

- 일반 지원 기종 / Standard supported cameras: **PC(테더) / PC(Tether)**
- **DC-L10: LUMIX Lab**

## 작동 관찰 / Observed behavior

다음 관찰은 **DC-S1RM2 / 펌웨어 1.5**에서 확인됐으며 Panasonic의 공식 카운터 정의가 아닙니다.
다른 모델에 그대로 적용하지 않습니다.

The following observations apply to **DC-S1RM2 / firmware 1.5** and are not official Panasonic definitions.
Do not extend them to other models.

| 동작 / Action | 변화 / Change |
|---|---|
| 기계식 셔터 1회 / One mechanical actuation | Shutter +1 |
| 전자식 촬영 / Electronic exposure | Shutter +0 |
| 전원 끌 때 셔터 닫기 / Power-off shutter CLOSE | Shutter +1 |
| 시험한 고해상도 촬영 설정 / Tested high-resolution settings | Shutter +0 |
| 센서 청소와 재시작 / Sensor cleaning and restart | Shutter +1 total observed |
| 픽셀 리프레시와 재시작, 전원 끌 때 셔터 닫기 / Pixel refresh and restart with power-off shutter CLOSE | Shutter +3 total; breakdown not isolated |
| 전원 끔 → 켬 / OFF → ON | Power-wake +1 |
| 절전 → 깨우기 / Sleep → wake | Power-wake +1; exact increment moment not isolated |
| USB만 재연결 / USB reconnect only | Power-wake +0 |

검증 데이터 갱신일 / Data revision: **2026-10-02**. [데이터 적용 / Apply updates](CAMERA_DATA.md)
