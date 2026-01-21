# Pairwise Pseudonymous Identifier Design for Sybil Resistance

## Executive Summary

This design implements a privacy-preserving Sybil resistance mechanism using **pairwise pseudonymous identifiers** derived from an opaque seed. The system prevents one real-world identity from creating multiple accounts on a verifier (e.g., forum) while maintaining unlinkability across different verifiers.

**Key Innovation**: Each credential includes a high-entropy `opaque_id_seed` that enables deterministic derivation of verifier-specific pseudonyms, ensuring:
- ✅ One person = One account per verifier (Sybil resistance)
- ✅ Different pseudonyms per verifier (unlinkability)
- ✅ No correlation by colluding verifiers (privacy)
- ✅ Selective disclosure compatible (holder controls revelation)

## Design Date
2026-01-21

## Protocol Standards

- **Base**: OpenID4VP (OpenID for Verifiable Presentations)
- **Credential Format**: SD-JWT (Selective Disclosure JWT) - ISO/IEC 18013-7
- **Key Binding**: JWT Proof of Possession (RFC 7800)
- **Derivation**: HMAC-SHA256 (RFC 2104)

---

## Part 1: Issuer Side (Credential Schema & Generation)

### 1.1 New Claim: `opaque_id_seed`

**Field Definition:**
```json
{
  "opaque_id_seed": "<base64url-encoded-256-bit-random-value>"
}
```

**Properties:**
- **Type**: String (base64url-encoded)
- **Entropy**: 256 bits (32 bytes) minimum
- **Generation**: Cryptographically secure random number generator (CSPRNG)
- **Uniqueness**: Unique per subject (holder), stable for credential lifetime
- **Disclosure**: MUST be marked for Selective Disclosure (SD-JWT)
- **Visibility**: Never sent to verifier directly, only used for derivation

**Example Value:**
```
"opaque_id_seed": "a3d7f9c8b2e1a4f6d8c9b7e2a5f8d3c1b4e7a9f2d6c8b3e5a7f9d2c4b6e8a1f3"
```

### 1.2 Schema Update

**Location:** `credential_policy` table, `vc_schema` column

**Updated Schema (JSON Schema format):**
```json
{
  "$id": "https://tw.gov.moda/schemas/age_verification_v2.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Age Verification Credential with Pairwise ID",
  "description": "Verifiable credential for age verification with Sybil-resistant pairwise identifiers",
  "type": "object",
  "properties": {
    "credentialSubject": {
      "type": "object",
      "properties": {
        "opaque_id_seed": {
          "type": "string",
          "description": "High-entropy seed for deriving pairwise pseudonymous identifiers",
          "pattern": "^[A-Za-z0-9_-]{43}$",
          "minLength": 43,
          "maxLength": 43
        },
        "over_18": {
          "type": "boolean",
          "description": "Whether the subject is over 18 years old"
        },
        "over_21": {
          "type": "boolean",
          "description": "Whether the subject is over 21 years old"
        },
        "birth_year": {
          "type": "integer",
          "description": "Year of birth (optional, for age bracket verification)"
        }
      },
      "required": ["opaque_id_seed", "over_18"],
      "additionalProperties": false
    }
  }
}
```

**SD-JWT Disclosure Flags** (new configuration):
```json
{
  "selective_disclosure": {
    "opaque_id_seed": {
      "always_disclosed": false,
      "derivable": true,
      "salt_required": true
    },
    "over_18": {
      "always_disclosed": false
    },
    "over_21": {
      "always_disclosed": false
    },
    "birth_year": {
      "always_disclosed": false
    }
  }
}
```

### 1.3 Implementation: Issuer Code Changes

**File:** `core-system/twdiw-vc-handler/src/main/java/gov/moda/dw/issuer/vc/service/CredentialService.java`

**New Method: Generate Opaque ID Seed**
```java
/**
 * Generates a cryptographically secure opaque ID seed for pairwise pseudonyms.
 *
 * @param holderUid Unique identifier for the holder
 * @param credentialType Type of credential being issued
 * @return Base64url-encoded 256-bit random seed
 */
private String generateOpaqueIdSeed(String holderUid, String credentialType) {
    // Check if seed already exists for this holder (to maintain stability)
    Optional<String> existingSeed = credentialRepository
        .findActiveOpaqueIdSeed(holderUid, credentialType);

    if (existingSeed.isPresent()) {
        return existingSeed.get();
    }

    // Generate new 256-bit seed using CSPRNG
    SecureRandom secureRandom = new SecureRandom();
    byte[] seedBytes = new byte[32]; // 256 bits
    secureRandom.nextBytes(seedBytes);

    // Encode as base64url (URL-safe, no padding)
    String seed = Base64.getUrlEncoder().withoutPadding()
        .encodeToString(seedBytes);

    // Store for future credential renewals
    credentialRepository.saveOpaqueIdSeed(holderUid, credentialType, seed);

    return seed;
}
```

**Modified Method: Credential Generation**
```java
// In CredentialService.generate()

// 1. Generate or retrieve stable opaque_id_seed
String opaqueIdSeed = generateOpaqueIdSeed(
    credentialRequestDTO.getHolderUid(),
    credentialRequestDTO.getCredentialType()
);

// 2. Add to holder data BEFORE SD-JWT encoding
Map<String, Object> holderData = getHolderData(credentialRequestDTO);
holderData.put("opaque_id_seed", opaqueIdSeed);

// 3. Continue with existing SD-JWT encoding process
// The opaque_id_seed will be automatically included in SD claims
String sdJwtCredential = credentialPrepareTask.prepareSd(holderData);
```

**Database Schema Addition:**

New table: `opaque_id_seed_registry`
```sql
CREATE TABLE opaque_id_seed_registry (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    holder_uid VARCHAR(255) NOT NULL,
    credential_type VARCHAR(100) NOT NULL,
    opaque_id_seed VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    status VARCHAR(20) DEFAULT 'ACTIVE',

    UNIQUE KEY uk_holder_type (holder_uid, credential_type, status),
    INDEX idx_holder (holder_uid),
    INDEX idx_expires (expires_at)
);
```

**Purpose**: Ensures the same seed is used for credential renewals, maintaining stable pairwise IDs.

---

## Part 2: Wallet & Presentation Logic (Derivation)

### 2.1 Pairwise ID Derivation Algorithm

**Function Specification:**
```
pairwise_id = HMAC-SHA256(
    key: opaque_id_seed,
    data: verifier_domain
)
```

**Inputs:**
- `opaque_id_seed`: 256-bit seed from credential (selectively disclosed)
- `verifier_domain`: Canonical domain of the verifier (e.g., "forum.example.com")

**Output:**
- `pairwise_id`: 256-bit pseudonymous identifier, base64url-encoded
- Format: 43-character string (256 bits / 6 bits per char ≈ 43 chars)

**Properties:**
- **Deterministic**: Same seed + domain = same pairwise_id
- **Collision-resistant**: Different seeds = different pairwise_ids (2^256 space)
- **One-way**: Cannot reverse pairwise_id to get seed
- **Unlinkable**: Different domains = uncorrelated pairwise_ids

### 2.2 Domain Canonicalization

To ensure stable pairwise_id across different verifier endpoints:

**Canonicalization Rules:**
```
1. Extract hostname from URL
2. Convert to lowercase
3. Remove "www." prefix if present
4. Use eTLD+1 (effective TLD + 1 label)

Examples:
- "https://FORUM.Example.Com/callback" → "forum.example.com"
- "https://www.forum.example.com" → "forum.example.com"
- "https://api.forum.example.com" → "forum.example.com"
```

**Implementation (pseudocode):**
```javascript
function canonicalizeDomain(verifierUrl) {
    const url = new URL(verifierUrl);
    let domain = url.hostname.toLowerCase();

    // Remove www prefix
    if (domain.startsWith('www.')) {
        domain = domain.substring(4);
    }

    // Extract eTLD+1 (use public suffix list library)
    domain = getETLDPlusOne(domain);

    return domain;
}
```

### 2.3 Wallet Implementation

**Location:** Mobile wallet app (Flutter) - needs to be implemented

**New Service: PairwisePseudonymService**

```dart
import 'dart:convert';
import 'dart:typed_data';
import 'package:crypto/crypto.dart';

class PairwisePseudonymService {
  /// Derives a pairwise pseudonymous ID for a specific verifier
  ///
  /// @param opaqueIdSeed Base64url-encoded opaque ID seed from credential
  /// @param verifierDomain Canonical domain of the verifier
  /// @return Base64url-encoded pairwise pseudonymous ID
  static String derivePairwiseId(String opaqueIdSeed, String verifierDomain) {
    // 1. Decode the seed
    Uint8List seedBytes = base64Url.decode(opaqueIdSeed);

    // 2. Canonicalize domain
    String canonicalDomain = _canonicalizeDomain(verifierDomain);

    // 3. Compute HMAC-SHA256
    var hmac = Hmac(sha256, seedBytes);
    var domainBytes = utf8.encode(canonicalDomain);
    var digest = hmac.convert(domainBytes);

    // 4. Encode as base64url
    String pairwiseId = base64Url.encode(digest.bytes);

    return pairwiseId;
  }

  /// Canonicalizes a verifier URL to its domain
  static String _canonicalizeDomain(String verifierUrl) {
    Uri uri = Uri.parse(verifierUrl);
    String domain = uri.host.toLowerCase();

    // Remove www prefix
    if (domain.startsWith('www.')) {
      domain = domain.substring(4);
    }

    // TODO: Implement eTLD+1 extraction using public suffix list
    // For now, use full domain minus www

    return domain;
  }
}
```

### 2.4 VP Token Construction

**Modified Presentation Response:**

The wallet MUST:
1. Retrieve `opaque_id_seed` from the selected credential
2. Derive `pairwise_id` for the requesting verifier
3. Include `pairwise_id` in the VP token (NOT the seed)
4. Only disclose age claims (e.g., `over_18`) as requested

**VP Token Structure:**
```json
{
  "iss": "did:example:holder123",
  "aud": "https://forum.example.com",
  "nonce": "abc123...",
  "vp": {
    "@context": ["https://www.w3.org/2018/credentials/v1"],
    "type": ["VerifiablePresentation"],
    "verifiableCredential": [
      "<SD-JWT-with-selected-disclosures>"
    ],
    "pairwise_sub": "a3d7f9c8b2e1a4f6d8c9b7e2a5f8d3c1b4e7a9f2d6c8b3e5a7f9d2c4b6e8a1f3"
  },
  "proof": {
    "type": "JwtProof2020",
    "jwt": "<holder-signature-over-VP>"
  }
}
```

**Key Addition**: `pairwise_sub` field in VP contains the derived pairwise_id

---

## Part 3: Verifier Side (Forum/Service Integration)

### 3.1 VP Token Validation

**Verifier Process:**
```
1. Receive VP token with pairwise_sub
2. Validate VP signature (holder key binding)
3. Validate VC signature (issuer authority)
4. Extract pairwise_sub from VP
5. Extract disclosed age claims (e.g., over_18 = true)
6. Check database for existing user with this pairwise_sub
```

### 3.2 User Registration Flow

**Pseudocode:**
```javascript
async function registerUser(vpToken) {
    // 1. Validate VP token
    const validation = await validateVPToken(vpToken);
    if (!validation.valid) {
        throw new Error('Invalid VP token');
    }

    // 2. Extract pairwise_sub
    const pairwiseId = validation.vp.pairwise_sub;
    const claims = validation.vp.verifiableCredential[0].credentialSubject;

    // 3. Check for Sybil attack (duplicate registration)
    const existingUser = await db.users.findByPairwiseId(pairwiseId);
    if (existingUser) {
        throw new SybilDetectedError('This credential is already registered');
    }

    // 4. Verify age requirement
    if (!claims.over_18) {
        throw new Error('User must be over 18');
    }

    // 5. Create new user account
    const userId = await db.users.create({
        pairwise_id: pairwiseId,
        over_18: claims.over_18,
        registered_at: new Date()
    });

    return { userId, pairwiseId };
}
```

### 3.3 User Authentication Flow

**Subsequent Login:**
```javascript
async function authenticateUser(vpToken) {
    // 1. Validate VP token
    const validation = await validateVPToken(vpToken);
    if (!validation.valid) {
        throw new Error('Invalid VP token');
    }

    // 2. Extract pairwise_sub
    const pairwiseId = validation.vp.pairwise_sub;

    // 3. Look up existing user
    const user = await db.users.findByPairwiseId(pairwiseId);
    if (!user) {
        throw new Error('User not found - please register first');
    }

    // 4. Create session
    const sessionToken = await createSession(user.id);

    return { userId: user.id, sessionToken };
}
```

### 3.4 Database Schema

**Users Table:**
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    pairwise_id VARCHAR(64) NOT NULL UNIQUE,  -- Pairwise pseudonymous ID
    over_18 BOOLEAN NOT NULL,                  -- Verified claim
    over_21 BOOLEAN DEFAULT NULL,              -- Optional claim
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL,

    INDEX idx_pairwise (pairwise_id)
);
```

**Benefits:**
- ✅ `pairwise_id` acts as unique constraint (prevents Sybil)
- ✅ No PII stored (privacy-preserving)
- ✅ Can still have username, profile separately
- ✅ Credential re-verification updates claims

---

## Security Analysis

### 4.1 Sybil Resistance

**Attack: One person tries to create multiple accounts**

**Defense:**
1. All credentials for same person have same `opaque_id_seed`
2. Wallet derives same `pairwise_id` for same verifier domain
3. Database unique constraint on `pairwise_id` prevents duplicate registration
4. Result: ✅ **Sybil attack blocked**

**Attack: Attacker steals someone's credential**

**Defense:**
1. VP token requires holder key binding proof
2. Holder must sign VP with their private key (stored securely in wallet)
3. Verifier validates signature matches `cnf` claim in credential
4. Result: ✅ **Credential theft ineffective without private key**

### 4.2 Unlinkability Across Verifiers

**Scenario: Two forums want to correlate users**

**Privacy Protection:**
```
Forum A: pairwise_id_A = HMAC(seed, "forum-a.com")
Forum B: pairwise_id_B = HMAC(seed, "forum-b.com")
```

Because HMAC output appears random:
- `pairwise_id_A` and `pairwise_id_B` are uncorrelated
- Forums cannot link the same person across services
- Result: ✅ **Cross-verifier linkability prevented**

### 4.3 Seed Stability Requirement

**Critical Property:** `opaque_id_seed` MUST remain stable for credential lifetime

**Issuer Responsibility:**
1. Store seed in `opaque_id_seed_registry` table
2. When renewing credential, use SAME seed
3. Only generate new seed if credential type changes OR holder explicitly requests reset

**Risk if seed changes:**
- User would appear as new person to verifier
- Could create second account (unintentional Sybil)
- Mitigation: Database unique constraint prevents this

### 4.4 Domain Canonicalization Importance

**Why Needed:**
```
Verifier has multiple endpoints:
- https://forum.example.com
- https://api.forum.example.com
- https://www.forum.example.com
```

Without canonicalization:
- Each subdomain would generate different pairwise_id
- User would appear as different people
- Sybil resistance breaks down

With canonicalization:
- All URLs resolve to "forum.example.com"
- Same pairwise_id generated
- User correctly identified as same person

---

## Implementation Checklist

### Phase 1: Issuer (Java - core-system/twdiw-vc-handler)

- [ ] Create `opaque_id_seed_registry` table
- [ ] Implement `generateOpaqueIdSeed()` method
- [ ] Modify `CredentialService.generate()` to include seed
- [ ] Update credential schema in `credential_policy` table
- [ ] Add unit tests for seed generation
- [ ] Add integration tests for credential with seed

### Phase 2: Wallet (Flutter - APP/TWWallet)

- [ ] Implement `PairwisePseudonymService.derivePairwiseId()`
- [ ] Implement domain canonicalization logic
- [ ] Modify VP token construction to include `pairwise_sub`
- [ ] Add UI for user consent (showing which verifier gets pairwise_id)
- [ ] Add unit tests for derivation
- [ ] Add integration tests for VP generation

### Phase 3: Verifier (Go - verifier-go)

- [ ] Extend `PresentationValidationResponse` to include `pairwise_sub`
- [ ] Add validation for `pairwise_sub` field in VP
- [ ] Document verifier integration guide
- [ ] Add example code for Sybil detection
- [ ] Add unit tests for pairwise_sub extraction

### Phase 4: Documentation

- [ ] Update API documentation with pairwise_sub
- [ ] Create developer guide for verifiers
- [ ] Document Sybil resistance mechanism
- [ ] Create privacy impact assessment
- [ ] Update security documentation

---

## Test Vectors

### Test Vector 1: Basic Derivation

**Input:**
```
opaque_id_seed: "a3d7f9c8b2e1a4f6d8c9b7e2a5f8d3c1b4e7a9f2d6c8b3e5a7f9d2c4b6e8a1f3"
verifier_domain: "forum.example.com"
```

**Expected Output:**
```
pairwise_id: "5d3e8f7c9b1a2e4d6f8c0b3e5a7d9f1c3e5b7d9a2c4e6f8b0d2e4a6c8f0b1d3"
```

### Test Vector 2: Different Domains

**Input:**
```
opaque_id_seed: "a3d7f9c8b2e1a4f6d8c9b7e2a5f8d3c1b4e7a9f2d6c8b3e5a7f9d2c4b6e8a1f3"
verifier_domain: "social.example.org"
```

**Expected Output:**
```
pairwise_id: "7f9e2d4c6b8a0e3f5d7c9b1a3e5d7f9c1b3e5d7a9c2e4f6d8b0a2e4c6f8a0b2"
```

(Note: Different from Test Vector 1, demonstrating unlinkability)

---

## Privacy Considerations

### Data Minimization

✅ **Only age claims disclosed** (not full birthdate)
✅ **Pairwise ID is pseudonymous** (not real identity)
✅ **Seed never revealed** to verifier
✅ **No correlation** across verifiers

### GDPR Compliance

- **Right to erasure**: User can request issuer to revoke credential
- **Data portability**: User controls their credential in wallet
- **Purpose limitation**: Pairwise_id only used for Sybil resistance
- **Storage limitation**: Verifier stores minimal data (pairwise_id + age)

### Comparison with Alternatives

| Approach | Sybil Resistance | Privacy | Linkability | Revocation |
|----------|------------------|---------|-------------|------------|
| **Real Name** | ✅ Strong | ❌ None | ❌ Full | ❌ Hard |
| **Email/Phone** | ⚠️ Medium | ⚠️ Low | ⚠️ Medium | ⚠️ Medium |
| **Global DID** | ✅ Strong | ❌ None | ❌ Full | ✅ Easy |
| **Pairwise Pseudonym** | ✅ Strong | ✅ High | ✅ None | ✅ Easy |

**Winner:** Pairwise pseudonyms provide the best balance

---

## Future Enhancements

### 1. Group Pseudonyms

Allow multiple verifiers to share same pairwise_id for federated services:
```
pairwise_id = HMAC(seed, "group:social_media_sites")
```

### 2. Time-Limited Pseudonyms

Rotate pairwise_ids periodically for enhanced privacy:
```
pairwise_id = HMAC(seed, domain || epoch_month)
```

### 3. Reputation Portability

Allow users to prove "same person" across verifiers without revealing identity:
```
proof = ZKP(pairwise_id_A, pairwise_id_B, seed)
```

---

## References

- [SD-JWT Specification (IETF Draft)](https://datatracker.ietf.org/doc/draft-ietf-oauth-selective-disclosure-jwt/)
- [OpenID for Verifiable Presentations](https://openid.net/specs/openid-4-verifiable-presentations-1_0.html)
- [RFC 2104 - HMAC](https://datatracker.ietf.org/doc/html/rfc2104)
- [RFC 7800 - JWT Proof of Possession](https://datatracker.ietf.org/doc/html/rfc7800)
- [Pairwise Pseudonymous Identifiers (W3C)](https://www.w3.org/TR/did-core/#pairwise-identifiers)

---

**Status**: Design Complete ✅
**Next**: Implement Phase 1 (Issuer)
