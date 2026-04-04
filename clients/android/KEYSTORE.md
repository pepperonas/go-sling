# Keystore

The release keystore is **not** checked into git.

## Location

- **Google Drive backup:** `My Drive/dev/keystore/go-sling-keystore/`
- **Local project:** `clients/android/release.jks` + `keystore.properties`

## Setup for building

Copy from Google Drive to the project:

```bash
cp "/Users/martin/My Drive/dev/keystore/go-sling-keystore/release.jks" clients/android/
cp "/Users/martin/My Drive/dev/keystore/go-sling-keystore/keystore.properties" clients/android/
```

Then build:

```bash
cd clients/android
./gradlew assembleRelease
```

The signed APK will be at `app/build/outputs/apk/release/app-release.apk`.

## Keystore Details

- **Alias:** gosling
- **Validity:** 10,000 days (~27 years)
- **Algorithm:** RSA 2048
- **DN:** CN=Martin Pfeffer, O=celox.io, L=Munich, C=DE
