# Viacom New Media Graphics Format Exporter

You'll never need this.

VNMGF is a proprietay sprite packing format used by Viacom circa 1995. To my knowledge, its only usage in the wild is via Beavis & Butthead's Virtual Insanity from that same year.

As for why I made this? I wanted to re-make the Hock-A-Loogie mini-game from Virtual Insanity in Godot for fun and as a learning exercise. So, I downloaded a copy of the original game from some dark, long forgotten corner of the internet. Inside were the sound files and some sort of .000 file that didn't match any file signature that I could find. I was confident the image data was packed in there, I just needed to figure out what format it was. I couldn't find any file formats whose signatures matched, alas.

Luckily, literally [one other person](http://www.shikadi.net/moddingwiki/User:Napalm) on the internet knows about these files and had already reverse engineered the format (http://www.shikadi.net/moddingwiki/Viacom_New_Media_Graphics_File_Format) back in 2015.

So, I spent some time converting the extractor to Go (based on the C code in the link above) to extract all the images.

The data does not seem to come out perfect. There are a few images with odd pixels missing near the right edge. But, it's 99% accurate and the issues are easily fixed manually.


The original `BBLOOGIE.000` file is included in the `data` dir along with the exported images. Knock yourself out.

# Additional Info

I am copying this information from Shikadi (see the link above) for the sake of preservation.

## VNM File Header
| Data Type | Description   |
| ----------| ------------- |
| UINT32LE  | Signature: File Signature (should match 0x564E4D1A aka "VNM\x1A")
| UINT32LE  | Flags: Possibly file information flags.
| UINT32LE  | Size: Size Of Data
| UINT32LE  | OffsetPal: Offset to Pal Data.
| UINT32LE  | OffsetUnknown1: Offset to currently Unknown Data.
| UINT32LE  | OffsetUnknown2: Offset to currently Unknown Data.
| UINT32LE  | OffsetImages: Offset to VNM Image index array.
| UINT32LE  | PalIndexFirst: First valid index in Color Palette.
| UINT32LE  | PalIndexSize: Count of Palette indexes stored in file at OffsetPal.
| UINT32LE  | ImageCount: Count of Images stored in file at OffsetImages.

## VNM Image Header
| Data Type | Description  |
| --------- | ------------ |
| UINT32LE  | Offset: Offset in file to where raw image data starts.
| UINT32LE  | Type: 0 = Bitmap, 1 = Sprite
| UINT32LE  | Width: Width of this Image
| UINT32LE  | Height: Height of this Image.
| UINT32LE  | XPos: XPos of this Image.
| UINT32LE  | YPos: YPos of this Image.

## File format description:

The file consists of the VNM File Header followed by VNM Image data.

Each VNM Image starts at a location in the file specified by the array found at
OffsetImages. This array is ImageCount in length and each entry specifies the
position in the file for a specific VNM Image.

A VNM Image consists of metadata (width, height, etc.) about a given image as
well as an Offset pointing to yet another array based index. This index has an
entry for each row in the image. Thus, to rebuild an image you traverse the row
index array and gather up the data starting from each position the index
specifies, grabbing Width bytes.

If the image is a Bitmap, you can put those bytes together and you're finished.

However, if the image is a Sprite, you must also expand the transparent
sections which are flagged via an escape code equal to 0x100 (256) minus the
image's width. When one of these bytes is encountered, you subtract the value
at the current pixel byte from 0x100 to determine how many transparent should
be added before moving on.

It looks something like:

```
F800                          19 FD00  18F700                              0F                 FB
F901                       18 19 FF00  18FF0117 19             F801        0E 10 FD00  0F
FA07                       19 18 18 17 18 19 17 19             F801        0E 10 FD00  0F
FA08                    19 17 19 19 17 19 18 18 19 FF00  19    FB05     0F 11 0F 11 0F 10
FB0B                    19 18 17 19 18 17 18 18 17 19 17 18    FB05     10 11 10 11 0E 11
FB0C                    18 18 19 19 18 17 18 19 17 19 17 19 18 FC04     11 0E 0F 10 0E        FF
FC0D                 19 17 19 18 19 19 17 18 19 17 18 17 19 18 FD05     10 0E 11 10 11 0F     FF
FC0D                 19 17 19 18 19 17 18 19 17 18 19 17 19 17 FD05     10 0F 11 11 10 11     FF
FC0D                 18 17 19 17 18 17 19 18 17 18 19 18 19 19 FC04     11 10 10 0F 10        FF
FC0E                 18 18 19 17 17 19 19 18 17 18 19 19 18 19 18 FD03  11 0F 0F 10           FE
FC0E                 19 18 19 18 19 18 19 18 17 18 19 18 17 18 19 FD02  11 11 10              FD
FC0E                 19 18 19 18 19 17 19 19 17 18 19 17 17 19 19 FD02  11 10 0F              FD
FC0E                 19 17 18 10 0F 18 19 17 18 19 19 18 17 19 19 FD02  11 0F 10              FD
FC0E                 18 17 19 0F 0E 0F 18 18 19 0F 10 18 17 19 18 FD02  11 0F 10              FD
FC0E                 19 17 19 0E 0E 0E 0E 0E 0E 0E 0E 0F 18 19 18 FD02  11 0E 10              FD
FC0E                 19 18 10 0E 0F 0F 0E 0E 0E 0E 0F 0F 10 17 19 FD02  11 0E 10              FD
FC0E                 18 19 0F 10 11 11 10 0E 0F 10 11 10 0F 17 19 FD02  11 0E 10              FD
FC0E                 17 19 0E 0F 10 10 11 0F 10 11 10 0F 0E 18 19 FD02  11 0E 10              FD
FC0E                 19 17 10 0E 0F 0E 0F 0E 10 0F 0E 0E 0E 19 18 FD02  11 0E 10              FD
FC0E                 18 11 0F 0E 0E 0F 10 0E 0E 10 0F 0E 0E 19 17 FD02  11 0E 10              FD
FB0D                 10 0F 0E 0E 10 11 0F 11 10 10 0E 0F 10 17 FD02     10 0E 11              FD
FB0C                 11 10 0F 0E 0E 0F 10 10 0F 0F 0E 0F 0FFC02         10 0E 11              FD
FA0B                    11 0F 0E 0E 0E 0F 0F 0F 0E 0E 0F 11FC02         0E 0F 11              FD
FA0B                    11 10 0F 0E 0F 20 10 20 0F 0E 0F 11FD02      10 0E 10                 FC
F90A                       10 0F 0E 10 20 20 20 10 0F 10 11FD02      0E 0F 11                 FC
F909                       11 0F 0E 11 15 16 15 11 0F 10FD02      10 0E 10                    FB
F909                       11 0F 0E 11 1C 1C 1C 11 0F 11FD02      0E 0F 11                    FB
F808                          10 0F 10 11 1C 11 10 0F 11FE02   10 0E 10                       FA
F807                          10 0F 0F 10 11 10 0F 10FE03   10 0E 0F 11                       FA
F807                          11 10 0F 0E 0F 0E 0F 11FF041B 10 0E 10 11                       FA
F70B                             11 10 0F 0E 0F 10 11 17 1B 10 0F 11                          F9
F70B                             11 10 10 10 10 11 17 1B 1A 1B 10 11                          F9
F60A                                11 10 0F 10 11 1C 1B 1A 1A 1B 1B                          F9
F70B                             1B 1B 10 10 10 1B 1C 1B 1A 1A 1A 1C                          F9
F90D                       1B 1B 1A 1A 1B 1B 1B 1A 1B 1A 1A 1A 1B 1C                          F9
FA0D                    1B 1A 1A 1A 1A 1A 1A 1A 1A 1A 1A 1A 1B 1C                             F8
FB0E                 1B 1A 1A 1A 1A 1A 1A 1A 1A 1A 1A 1A 1A 1B 1C                             F8
FC0E              1B 1A 1A 1A 1B 1A 1A 1A 1A 1A 1A 1A 1A 1B 1C                                F7
FC0E              1B 1A 1A 1B 1C 1B 1A 1A 1A 1A 1A 1A 1B 1B 1C                                F7
FD0F           10 10 1B 1B 1C 1B 1B 1A 1A 1A 1A 1A 1A 1B 1B 1C                                F7
FD0F           10 0F 10 1C 1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1B 1C                                F7
FE03        10 0E 10 11FF0A1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
FE02        10 0F 11   FE0A1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
FF03     10 0E 10 11   FE0A1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
FF02     10 0F 11   FD0A   1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
FF02     10 10 11   FD0A   1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
02    10 0F 10      FC0A   1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
02    10 0E 11      FC0A   1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
02    10 0E 11      FC0A   1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
02    10 0E 11      FC0A   1C 1B 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
02    10 0E 11      FD0B1C 1B 1A 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
02    10 0F 11      FD0B1C 1B 1A 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
02    10 10 11      FD0B1C 1B 1A 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
03    10 0F 0F 10   FE0B1C 1B 1A 1A 1A 1A 1A 1A 1A 1A 1B 1C                                   F6
04    11 0E 10 0F 10FF0B1C 1B 1B 1A 1A 1A 1A 1A 1A 1B 1B 1C                                   F6
03    11 0E 0F 11   FE0B1C 1C 1C 1B 1B 1B 1B 1B 1B 1C 1C 1C                                   F6
03    11 0F 0E 10   FE0B22 22 22 21 21 21 21 21 21 21 22 22                                   F6
FF02     11 11 11   FE0B22 21 21 20 20 20 20 20 20 20 21 22                                   F6
FA0B                    22 21 20 20 20 20 20 20 20 20 21 22                                   F6
FA0B                    22 21 20 20 20 21 20 20 20 20 21 22                                   F6
FA0B                    22 21 20 20 21 22 21 20 20 20 21 22                                   F6
FA0B                    22 21 20 20 21 22 22 21 20 20 21 22                                   F6
FA0B                    22 21 21 21 21 22 22 21 21 21 21 22                                   F6
F909                       11 10 0F 10 11 11 10 0F 10 11                                      F5
F909                       11 0F 0E 0F 11 11 0F 0E 0F 11                                      F5
F909                       11 0F 0E 0F 11 11 0F 0E 0F 11                                      F5
F909                       11 0F 0E 0F 11 11 0F 0E 0F 11                                      F5
F909                       11 0F 0E 0F 11 11 0F 0E 0F 11                                      F5
F909                       11 10 0E 0F 11 11 0F 0E 10 11                                      F5
F807                          11 0E 0F 11 11 0F 0E 11                                         F4
F807                          11 0E 0F 11 11 0F 0E 11                                         F4
F807                          11 0E 0F 11 11 0F 0E 11                                         F4
F807                          11 0E 0F 11 11 0F 0E 11                                         F4
F807                          11 0E 10 11 11 10 0E 11                                         F4
F802                          16 16 16  FE02 16 16 16                                         F4
F802                          16 15 16  FE02 16 15 16                                         F4
F802                          16 15 16  FE02 16 15 16                                         F4
F802                          16 15 16  FE02 16 15 16                                         F4
F807                          1C 1C 1C 1C 1C 1C 1C 1C                                         F4
FA0B                    1C 1B 1A 1B 1B 1C 1C 1B 1B 1A 1B 1C                                   F6
FB0D                 1C 1B 1A 1B 1B 1C 1C 1C 1C 1B 1B 1A 1B 1C                                F7
FB04                 1C 1C 1C 1C 1C     FC04    1C 1C 1C 1C 1C                                F7
```
