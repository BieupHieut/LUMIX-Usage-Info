# Lumerian Brand & Character Guide

## Brand role

Lumerian is the friendly photographer who guides users through LUMIX camera usage information. The name appears as the product-facing brand while `LUMIX Usage Info` remains the functional description of the tool.

## Character identity

- Character: small cream and warm golden-tan puppy with long floppy ears
- Face: broad white forehead blaze and muzzle, round glossy black eyes, black nose and pink cheeks
- Role: photographer, technical guide and calm troubleshooting companion
- Signature items: bright red hoodie with cream drawstrings and a black mirrorless camera
- Visual accent: deep red primary actions and shutter cues, deep blue USB / connected states, and neutral borders on near-black surfaces
- Personality: curious, capable, warm and reassuring

The floppy-ear silhouette, white facial blaze, pink cheeks, red hoodie and camera must remain consistent across every state. Lumerian should look concerned rather than frightened in error screens.

## Product sprite states

| Asset | Purpose | Expression / action |
| --- | --- | --- |
| `mascot_default.bmp` | Main and connected views | Friendly smile, camera held at chest |
| `mascot_info.bmp` | Connection guide, technical help and project information | Official 2D sticker holding a camera-history card |
| `mascot_export.bmp` | PNG and report export | Presents a finished photo with gold sparkles |
| `mascot_error.bmp` | Disconnected or error state | Calm concern with amber warning cue |

Runtime BMP assets are 128×128 with a `#121214` background so their edges blend into the Windows app. The official `Lumerian_In_App.png` is used unchanged as the default high-resolution master; matching information, export and error masters are kept in `branding/lumerian_master`.

In beta.14, the character stays in the header at the same 128×128 slot on every page. Only its pose and style change. Language controls occupy a separate fixed area to its left. The information state uses the official `Lumerian_2D_Sticker.png`, copied as `branding/lumerian_master/lumerian_info_2d_master.png`; the previous 3D information master is retained. Connected, export and error states continue using the 3D illustrations.

## Product UI palette

- Canvas: `#121214`; card surface: `#1C1C1F`; routine border: `#47474D`.
- Deep red `#B54754`: connection CTA; `#A5444D`: shutter border and recoverable warning cues. Light red `#F1A3A9` is used for text.
- Deep blue `#386CA8`: USB / connected border and power/wake border. Light blue `#A7C8F0` is used for text.
- Text remains white or muted gray. The red accents echo Lumerian's hoodie sparingly; blue identifies the USB and connected-camera information.

## Product wording

- Primary brand: **Lumerian**
- Functional descriptor: **LUMIX Camera Usage Reader**
- Window title: **Lumerian · LUMIX Usage Info**
- Brand line: **Lumerian · Good Photos, Better Days!**

The tool remains an unofficial community project and does not claim affiliation with or endorsement by Panasonic.

## Creator identity — Bieup_Hieut

The left half of Developer Credits uses the supplied `Bieup_Hieut_In_App.png` character: cream hoodie with the ㅂㅎ mark, denim shorts, a friendly smile and a thumbs-up pose. This is the creator character; Lumerian remains the puppy character guiding the camera product.

The unchanged original is kept as `branding/creator_character_original.png`. The UI embeds `source/assets/creator_character.bmp` (264×240, aspect ratio preserved, transparency composited over the #222226 card). This conversion supports the existing native bitmap renderer. No generated replacement art was used. The earlier creator logo master remains as historical branding material.
