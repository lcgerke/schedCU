# Security Gaps & Fixes

Critical security issues that must be addressed before production deployment and before v2 implementation.

---

## 🔴 CRITICAL: Admin Endpoints Unprotected

**Impact**: Anyone with network access can:
- Trigger Amion scraping (hijack imports)
- Import arbitrary ODS files (inject data)
- Promote versions to production (schedule disruption)
- Export HTML (information disclosure)

**Locations**: `src/main/java/org/hospital/radiology/schedule/api/AdminResource.java`

### Current Code

```java
@Path("/api/admin")
@RestController
public class AdminResource {

    // ❌ Line 92
    @PermitAll  // TODO: Remove after testing
    @PostMapping("/import/amion")
    public Response scrapeAmion() {
        log.info("Starting Amion scrape...");
        ValidationResult result = amionImportService.scrapeAndImport(6);
        return Response.ok(new ApiResponse<>(result)).build();
    }

    // ❌ Line 118
    @PermitAll  // TODO: Remove after testing
    @PostMapping("/scrape-html")
    public Response scrapeHtml(@FormParam("month") String month) {
        YearMonth ym = YearMonth.parse(month);
        String html = scraper.scrapeAmion(ym);
        return Response.ok(new HtmlResponse(html)).build();
    }

    // ❌ Line 149
    @PermitAll  // TODO: Remove after testing
    @PostMapping("/import-html")
    public Response importHtml(@FormParam("html") String html) {
        ScrapeBatch batch = amionImportService.importHtml(html);
        return Response.ok(batch).build();
    }

    // ❌ Line 197
    @PermitAll  // TODO: Remove after testing
    @PostMapping("/workflow")
    public Response executeWorkflow(@FormParam("file") InputStream file) {
        ScheduleVersion version = orchestrator.executeCompleteWorkflow(file, ...);
        return Response.ok(version).build();
    }

    // ❌ Line 303
    @PermitAll  // TODO: Remove after testing
    @PostMapping("/resolve-coverage")
    public Response resolveCoverage(@FormParam("batchId") String batchId) {
        coverageResolution.resolveCoverage(batchId);
        return Response.ok().build();
    }

    // ✅ Line 120 (correctly protected)
    @RolesAllowed("ADMIN")
    @PostMapping("/promote/{versionId}")
    public Response promoteVersion(@PathParam("versionId") Long versionId) {
        // Only ADMIN can promote
        ...
    }
}
```

### Attack Scenarios

**Scenario 1: Data Injection**
```bash
# Anyone can upload malicious ODS file
curl -X POST http://localhost:8081/api/admin/import/ods \
  -F "file=@malicious.ods"

# Malicious file could:
# - Add fake shifts to critical dates
# - Assign impossible coverage (1 person to 5 shifts)
# - Inject XXE attack during parsing
```

**Scenario 2: Schedule Hijacking**
```bash
# Import fake Amion data
curl -X POST http://localhost:8081/api/admin/scrape-html \
  -d "month=2024-11"

# Replace current schedule with fake one
# Physicians show up for shifts that don't exist
# Hospital loses coverage on critical dates
```

**Scenario 3: Privilege Escalation (if any)**
```bash
# If workflow endpoint creates admin users or assigns roles
curl -X POST http://localhost:8081/api/admin/workflow \
  -F "file=@payload.ods"

# Could potentially elevate own permissions
```

### Fix (Immediate)

Replace all `@PermitAll` with `@RolesAllowed("ADMIN")`:

```java
@Path("/api/admin")
@RestController
public class AdminResource {
    @RolesAllowed("ADMIN")  // ✅ Fixed
    @PostMapping("/import/amion")
    public Response scrapeAmion() { ... }

    @RolesAllowed("ADMIN")  // ✅ Fixed
    @PostMapping("/scrape-html")
    public Response scrapeHtml(@FormParam("month") String month) { ... }

    @RolesAllowed("ADMIN")  // ✅ Fixed
    @PostMapping("/import-html")
    public Response importHtml(@FormParam("html") String html) { ... }

    @RolesAllowed("ADMIN")  // ✅ Fixed
    @PostMapping("/workflow")
    public Response executeWorkflow(@FormParam("file") InputStream file) { ... }

    @RolesAllowed("ADMIN")  // ✅ Fixed
    @PostMapping("/resolve-coverage")
    public Response resolveCoverage(@FormParam("batchId") String batchId) { ... }
}
```

### Add Tests

```java
@QuarkusTest
@TestSecurity(user = "user")
public class AdminEndpointSecurityTest {

    @Test
    public void testAdminEndpointsRequireAuth() {
        // No token → 401
        given()
            .when().post("/api/admin/import/amion")
            .then().statusCode(401);
    }

    @Test
    @TestSecurity(user = "user", roles = "USER")
    public void testAdminEndpointsRequireAdminRole() {
        // Token with USER role → 403
        given()
            .auth().oauth2(userToken)
            .when().post("/api/admin/import/amion")
            .then().statusCode(403);
    }

    @Test
    @TestSecurity(user = "admin", roles = "ADMIN")
    public void testAdminEndpointsAllowAdminRole() {
        // Token with ADMIN role → allowed
        given()
            .auth().oauth2(adminToken)
            .when().post("/api/admin/import/amion")
            .then().statusCode(200);
    }
}
```

---

## 🟠 HIGH: Hardcoded Database Credentials

**Severity**: HIGH
**Impact**: Credentials in Git history, exposed in Docker logs
**File**: `src/main/resources/application.properties`, line 12

### Current Code

```properties
quarkus.datasource.username=postgres
quarkus.datasource.password=postgres  # ❌ Hardcoded
```

### Problem

```bash
# Credential visible in:
1. Source code repositories
2. Docker images
3. Build logs
4. Environment inspection
5. Configuration dumps
```

### Fix

```properties
# application.properties
quarkus.datasource.username=${DB_USER:postgres}
quarkus.datasource.password=${DB_PASSWORD}  # ← Required env var, no default
```

Environment setup:
```bash
# .env file (never commit)
DB_USER=postgres
DB_PASSWORD=your-secure-password-here

# Or set in Docker
docker run -e DB_PASSWORD=secret ...

# Or in Kubernetes
kubectl create secret generic db-credentials \
  --from-literal=password=secret
```

**Verify**: Test that application fails to start without env var
```bash
# Should fail
java -jar app.jar

# Should succeed
export DB_PASSWORD=test
java -jar app.jar
```

---

## 🟠 HIGH: No Input Validation on File Uploads

**Severity**: HIGH
**Impact**: XXE attacks, malicious file injection, DoS
**File**: `src/main/java/org/hospital/radiology/schedule/api/AdminResource.java`

### Attack Scenarios

**Scenario 1: XXE (XML External Entity)**
```xml
<!-- malicious.ods (actually ODS is ZIP + XML) -->
<?xml version="1.0"?>
<!DOCTYPE foo [
  <!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<spreadsheet>&xxe;</spreadsheet>
```

**Scenario 2: Zip Bomb**
```bash
# Create file that expands to 10GB when unpacked
dd if=/dev/zero bs=1M count=10000 | gzip > bomb.ods

# Upload → server hangs unpacking
curl -F "file=@bomb.ods" http://localhost:8081/api/admin/import/ods
```

**Scenario 3: Wrong File Type**
```bash
# Upload JPG, Excel, or PDF instead of ODS
# Parser might crash or misinterpret
```

### Fix

```java
@RolesAllowed("ADMIN")
@PostMapping("/import/ods")
public Response importODS(
    @FormParam("file") InputStream fileStream,
    @FormParam("filename") String filename) {

    try {
        // 1. Validate filename
        if (filename == null || !filename.toLowerCase().endsWith(".ods")) {
            return Response.status(400)
                .entity(new ApiError("INVALID_FILE_TYPE", "Only .ods files allowed"))
                .build();
        }

        // 2. Validate file size (max 50MB)
        long maxSize = 50 * 1024 * 1024;
        long fileSize = fileStream.available();
        if (fileSize > maxSize) {
            return Response.status(413)
                .entity(new ApiError("FILE_TOO_LARGE",
                    "File exceeds maximum size of 50MB"))
                .build();
        }

        // 3. Copy to temp file with timeout protection
        Path tempFile = createTempFile(filename);
        try (InputStream is = fileStream;
             OutputStream os = Files.newOutputStream(tempFile)) {

            byte[] buffer = new byte[8192];
            long totalRead = 0;
            int bytesRead;

            while ((bytesRead = is.read(buffer)) != -1) {
                totalRead += bytesRead;
                if (totalRead > maxSize) {
                    Files.delete(tempFile);
                    return Response.status(413)
                        .entity(new ApiError("FILE_TOO_LARGE", "..."))
                        .build();
                }
                os.write(buffer, 0, bytesRead);
            }
        }

        // 4. Validate ZIP structure (ODS is ZIP)
        if (!isValidODSFile(tempFile)) {
            Files.delete(tempFile);
            return Response.status(400)
                .entity(new ApiError("INVALID_ODS_FILE", "File is not valid ODS format"))
                .build();
        }

        // 5. Parse with XXE protection
        DirectODSParser parser = new DirectODSParser();
        parser.disableXXE();  // Ensure XXE is disabled
        ScheduleVersion version = parser.parse(tempFile);

        // 6. Cleanup
        Files.delete(tempFile);

        return Response.ok(new ApiResponse<>(version)).build();

    } catch (IOException e) {
        log.error("File upload error", e);
        return Response.status(500)
            .entity(new ApiError("UPLOAD_FAILED", e.getMessage()))
            .build();
    }
}

private boolean isValidODSFile(Path file) {
    try (ZipFile zip = new ZipFile(file.toFile())) {
        // Valid ODS must contain content.xml
        return zip.getEntry("content.xml") != null;
    } catch (IOException e) {
        return false;  // Not valid ZIP/ODS
    }
}

private Path createTempFile(String filename) throws IOException {
    // Use temp directory with proper permissions
    Path tempDir = Paths.get(System.getProperty("java.io.tmpdir"), "schedjas-uploads");
    Files.createDirectories(tempDir);

    // Generate random name to prevent path traversal
    String safeName = UUID.randomUUID().toString() + ".tmp";
    Path tempFile = tempDir.resolve(safeName);

    // Set restrictive permissions (Java doesn't fully support, but try)
    Set<PosixFilePermission> perms = PosixFilePermissions.fromString("rw-------");
    return Files.createFile(tempFile,
        PosixFilePermissions.asFileAttribute(perms));
}
```

### Verify XXE Protection

Test that XXE attack is blocked:
```java
@Test
public void testXXEAttackIsBlocked() throws IOException {
    String xxePayload = """
        <?xml version="1.0"?>
        <!DOCTYPE foo [
          <!ENTITY xxe SYSTEM "file:///etc/passwd">
        ]>
        <spreadsheet>&xxe;</spreadsheet>
    """;

    // Create malicious ODS with XXE
    File malicious = createODSWithXXE(xxePayload);

    // Attempt upload
    Response response = given()
        .auth().oauth2(adminToken)
        .multiPart("file", malicious)
        .when()
        .post("/api/admin/import/ods");

    // Should either fail or strip XXE
    assertTrue(response.statusCode() == 400 || !responseContains("/etc/passwd"));
}
```

---

## 🟠 HIGH: Hardcoded Amion File ID

**Severity**: HIGH
**Impact**: If file ID changes, system breaks; if leaked, Amion data could be scraped by unauthorized parties
**File**: `src/main/resources/application.properties`, line 56

### Current Code

```properties
scraping.amion.file-id=!1854eed1hnew_28821_6
```

### Questions

1. What is this ID? (Patient data identifier? Hospital identifier?)
2. Is this sensitive? (If leaked, could unauthorized parties scrape Amion?)
3. How often does it change?
4. What happens if it expires?

### Fix

Move to environment variable with validation:

```properties
# application.properties
scraping.amion.file-id=${AMION_FILE_ID:}
scraping.amion.file-id-required=true
```

Validate on startup:
```java
@ApplicationScoped
public class AmionConfigValidator {
    @ConfigProperty(name = "scraping.amion.file-id")
    Optional<String> fileId;

    @PostConstruct
    void validate() {
        if (fileId.isEmpty()) {
            throw new IllegalStateException(
                "AMION_FILE_ID environment variable is required");
        }

        if (!fileId.get().startsWith("!")) {
            throw new IllegalStateException(
                "AMION_FILE_ID should start with '!' (invalid format)");
        }

        if (fileId.get().length() < 10) {
            throw new IllegalStateException(
                "AMION_FILE_ID appears too short (invalid format)");
        }

        LOG.infof("Amion file ID configured: %s***",
            fileId.get().substring(0, 4));
    }
}
```

---

## 🟡 MEDIUM: No Rate Limiting on Login

**Severity**: MEDIUM
**Impact**: Brute force attacks possible
**File**: `src/main/java/org/hospital/radiology/schedule/api/AuthResource.java`

### Fix

```java
@Path("/api/auth")
public class AuthResource {
    private static final int MAX_LOGIN_ATTEMPTS = 5;
    private static final int LOCKOUT_MINUTES = 15;

    private final Map<String, LoginAttempt> loginAttempts = new ConcurrentHashMap<>();

    @PostMapping("/login")
    public Response login(LoginRequest request) {
        String email = request.email;

        // Check if account is locked
        if (isAccountLocked(email)) {
            return Response.status(429)  // Too Many Requests
                .entity(new ApiError("ACCOUNT_LOCKED",
                    "Too many failed login attempts. Try again in 15 minutes."))
                .build();
        }

        try {
            User user = User.find("email", email)
                .firstResultOptional()
                .orElseThrow(() -> new AuthException("Invalid credentials"));

            if (!PasswordService.verify(request.password, user.passwordHash)) {
                recordFailedAttempt(email);
                throw new AuthException("Invalid credentials");
            }

            // Success - clear attempts
            loginAttempts.remove(email);

            // Generate token
            String token = jwtService.generateToken(user, Duration.ofHours(8));

            // Audit
            AuditLog.create(user, "LOGIN", "Success", "");

            return Response.ok(new LoginResponse(token, user.email)).build();

        } catch (AuthException e) {
            AuditLog.create(email, "LOGIN", "Failed", e.getMessage());
            return Response.status(401)
                .entity(new ApiError("INVALID_CREDENTIALS", "Invalid email or password"))
                .build();
        }
    }

    private boolean isAccountLocked(String email) {
        LoginAttempt attempt = loginAttempts.get(email);
        if (attempt == null) return false;

        if (attempt.failureCount < MAX_LOGIN_ATTEMPTS) return false;

        long lockoutExpiry = attempt.lastFailureTime + (LOCKOUT_MINUTES * 60 * 1000);
        if (System.currentTimeMillis() > lockoutExpiry) {
            // Lockout expired
            loginAttempts.remove(email);
            return false;
        }

        return true;
    }

    private void recordFailedAttempt(String email) {
        loginAttempts.computeIfAbsent(email, k -> new LoginAttempt())
            .recordFailure();

        if (loginAttempts.get(email).failureCount >= MAX_LOGIN_ATTEMPTS) {
            AuditLog.create(null, "LOGIN_BLOCKED",
                "Too many failed attempts for " + email, "");
        }
    }

    private static class LoginAttempt {
        int failureCount = 0;
        long lastFailureTime = 0;

        void recordFailure() {
            failureCount++;
            lastFailureTime = System.currentTimeMillis();
        }

        void reset() {
            failureCount = 0;
            lastFailureTime = 0;
        }
    }
}
```

---

## 🟡 MEDIUM: Database Passwords in Logs

**Severity**: MEDIUM
**Impact**: Credentials leaked in application logs

### Fix

Disable SQL logging in production:
```properties
# application.properties
quarkus.hibernate-orm.log.sql=true

# Override in production
%prod.quarkus.hibernate-orm.log.sql=false
```

Do not log connection URLs:
```java
// ❌ Bad
LOG.infof("Connecting to database: %s", connectionUrl);

// ✅ Good
LOG.infof("Connecting to database: %s", maskConnectionUrl(connectionUrl));

private String maskConnectionUrl(String url) {
    return url.replaceAll("password=[^&;]+", "password=***");
}
```

---

## 🟡 MEDIUM: No CSRF Protection

**Severity**: MEDIUM
**Impact**: If accessing from browser, form submissions could be forged from malicious site
**Mitigation**: The system uses JWT which provides inherent CSRF protection (tokens required in headers, not cookies)

However, if cookies are used, add CSRF token:
```java
@PostMapping("/import/ods")
public Response importODS(
    @FormParam("file") InputStream file,
    @FormParam("_csrf") String csrfToken) {  // Add CSRF token validation

    if (!csrfValidator.isValid(csrfToken)) {
        return Response.status(403)
            .entity(new ApiError("INVALID_CSRF_TOKEN", "..."))
            .build();
    }

    // Process...
}
```

**Better**: Stick with JWT in Authorization header (current approach is fine).

---

## Summary: Security Fix Priority

| Issue | Severity | Effort | Risk if Ignored |
|-------|----------|--------|-----------------|
| Admin endpoints unprotected | 🔴 Critical | 1 hour | Production takeover |
| Hardcoded DB password | 🟠 High | 1 hour | Credential leak |
| No file upload validation | 🟠 High | 4 hours | XXE/injection attacks |
| Hardcoded Amion file ID | 🟠 High | 2 hours | Config leak, service disruption |
| No rate limiting | 🟡 Medium | 2 hours | Brute force attacks |
| Passwords in logs | 🟡 Medium | 1 hour | Credential leak |
| CSRF protection | 🟡 Medium | 2 hours | Form hijacking (low risk with JWT) |

**Total estimated effort**: ~14 hours for comprehensive security hardening

**Minimum for production**: 4 hours (fix critical + high issues)

### Deployment Checklist

- [ ] Remove all `@PermitAll` annotations from admin endpoints
- [ ] Move all credentials to environment variables
- [ ] Add file upload validation (size, type, XXE protection)
- [ ] Move Amion file ID to environment
- [ ] Add login rate limiting
- [ ] Disable SQL logging in production
- [ ] Add security tests to verify endpoints are protected
- [ ] Run security audit/code review before production
- [ ] Add OWASP dependency check to build pipeline

---

## Testing Security Fixes

```java
@QuarkusTest
public class SecurityFixesTest {
    @Test
    public void allAdminEndpointsRequireAuth() {
        // List of admin endpoints
        String[] endpoints = {
            "/api/admin/import/amion",
            "/api/admin/scrape-html",
            "/api/admin/import-html",
            "/api/admin/workflow",
            "/api/admin/resolve-coverage"
        };

        for (String endpoint : endpoints) {
            given()
                .when().post(endpoint)
                .then()
                .statusCode(401)  // Unauthorized
                .statusCode(403); // Forbidden (no token)
        }
    }

    @Test
    public void fileUploadRejectsLargeFiles() {
        // Create 100MB file
        File large = createTempFile(100 * 1024 * 1024);

        Response response = given()
            .auth().oauth2(adminToken)
            .multiPart("file", large)
            .when()
            .post("/api/admin/import/ods");

        assertEquals(413, response.statusCode());  // Payload Too Large
    }

    @Test
    public void fileUploadRejectsWrongFileType() {
        File jpgFile = new File("test.jpg");

        Response response = given()
            .auth().oauth2(adminToken)
            .multiPart("file", jpgFile)
            .when()
            .post("/api/admin/import/ods");

        assertEquals(400, response.statusCode());  // Bad Request
    }
}
```

This document should be the first thing the team addresses before production deployment.
