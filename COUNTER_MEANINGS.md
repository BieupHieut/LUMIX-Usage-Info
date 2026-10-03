# Counter meanings and evidence

1 = power/wake; 2 = shutter. Verified only for exact pairs in camera_profiles.json. Others display estimated names with original field positions.

3 = Flash Activations (estimated); 4 = Power-saving Events (estimated). Previous beta.6 source recorded STBCNT / PSVCNT as legacy candidates at these positions. The expansion into modern field meanings is a **hypothesis**, not a measured or official index mapping. Panasonic's [Data Act notice](https://www.panasonic.com/uk/consumer/eu-data-act/dsc.html) lists flash/power-saving history as data types for the relevant cameras but does not specify their USB offsets or confirm those candidate positions.

5–7 = unidentified. No invented sensor-cleaning, pixel-refresh or abnormal-shutdown names. The existing controlled S1RM2 observations did not identify a separate cleaning count. No claim that raw counter 4 diagnoses failed shutdowns.

The current output retains every original numeric value. Estimated names do not turn a profile into VERIFIED. Detailed shutter behavior remains scoped to DC-S1RM2 / firmware 1.5 in KNOWN_COUNTER_BEHAVIOR.md.
