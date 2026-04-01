# CryptoBox

Shared Swift package providing XSalsa20-Poly1305 authenticated encryption with Curve25519 key agreement.

## Dependencies

- `jedisct1/swift-sodium` (revision-pinned) — libsodium `crypto_box` and `crypto_secretbox`

## Source

```
├── Package.swift
├── Sources/CryptoBox/XSalsa20Poly1305Box.swift
└── Tests/CryptoBoxTests/XSalsa20Poly1305BoxTests.swift
```

## Public API

```swift
enum XSalsa20Poly1305Box {
    static func seal(plaintext: Data, recipientPublicKey: Data) throws
        -> (ephemeralPublicKey: Data, nonce: Data, ciphertext: Data)

    static func open(ciphertext: Data, nonce: Data, sharedSecret: Data) throws
        -> Data

    static func open(ciphertext: Data, nonce: Data,
                     ephemeralPublicKey: Data, recipientSecretKey: Data) throws
        -> Data
}

enum CryptoBoxError: Error {
    case invalidNonce, invalidCiphertext, authenticationFailed
    case invalidRecipientPublicKey, invalidEphemeralPublicKey, randomGenerationFailed
}
```

## Internals

- 32-byte keys, 24-byte nonces
- libsodium `crypto_box_easy` for authenticated public-key encryption
- libsodium `crypto_secretbox_open_easy` plus `crypto_core_hsalsa20` for shared-secret opens
- Ephemeral Curve25519 keypair generated per `seal` call
- Package remains a thin compatibility wrapper around the existing API

## Conventions

- All methods are static on an enum (no instances)
- Private helper functions, pure functional style
- Platforms: iOS 17+, macOS 13+
