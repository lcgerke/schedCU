# v1 Security Fixes Implementation Guide

**Duration**: Week 1, 6-8 hours (parallel with Phase 0)
**Owner**: API & Security Engineer
**Status**: Ready to execute after Week 0 completes
**Goal**: Make v1 production-ready by removing security gaps before v2 development

---

## Executive Summary

The Hospital Radiology Schedule System v1 (Java/Quarkus) has 4 known security gaps that must be fixed immediately. This is a **parallel track** during Week 1 — while v2 Phase 0 begins, 1 person fixes v1 security issues and deploys to production.

**Why this matters:**
- Removes security risk from production system
- Eliminates urgency pressure on v2 development
- Gives team confidence v1 is stable while rewrite proceeds
- Demonstrates commitment to security-first approach

**Timeline**: 6-8 hours total
- Security Fix 1: 2 hours
- Security Fix 2: 2 hours
- Security Fix 3: 2 hours
- Testing & Deployment: 2 hours

---

## Security Fix 1: Remove @PermitAll Bypass on Admin Endpoints

**Duration**: 2 hours
**Risk Level**: High
**Impact**: Admin endpoints are currently accessible without proper authentication

### Problem Statement

The v1 codebase has `@PermitAll` annotations on administrative endpoints that were intended for development/testing but left in production. This allows:
- Unauthenticated users to create/modify schedules
- Unauthenticated access to admin dashboard
- Potential data manipulation without audit trail

Example (current vulnerable code):
```java
@Path("/admin/schedules")
@PermitAll  // ⚠️ PROBLEM: No authentication required
public class ScheduleAdminResource {
    @POST
    public Response createSchedule(ScheduleDTO schedule) {
        // Creates schedule without authentication check
        return scheduleService.save(schedule);
    }
}
```

### Solution

Replace `@PermitAll` with proper role-based access control using `@RolesAllowed`.

#### Step 1: Identify all @PermitAll annotations (20 minutes)

```bash
cd /path/to/schedcu-v1
grep -r "@PermitAll" --include="*.java" | tee /tmp/permitall_audit.txt

# Expected: 3-5 endpoints in admin/* and api/* paths
```

Document findings in `/tmp/permitall_audit.txt`:
- File path
- Class name
- Method name
- Endpoint path

#### Step 2: Add @RolesAllowed annotations (40 minutes)

For each `@PermitAll` endpoint:

1. Remove `@PermitAll` annotation
2. Add `@RolesAllowed({"ADMIN", "SUPERVISOR"})`
3. Verify method uses authentication context

Example fix:
```java
@Path("/admin/schedules")
@RolesAllowed({"ADMIN", "SUPERVISOR"})  // ✓ FIXED: Role-based access
public class ScheduleAdminResource {
    @Inject
    private SecurityContext securityContext;

    @POST
    public Response createSchedule(ScheduleDTO schedule) {
        String user = securityContext.getUserPrincipal().getName();

        // Add audit log
        auditLog.record("SCHEDULE_CREATE", user, schedule.getId());

        return scheduleService.save(schedule);
    }
}
```

#### Step 3: Test role enforcement (20 minutes)

Create test to verify fix:
```java
@Test
public void testAdminEndpointRequiresRole() {
    // Arrange: Create request without authentication
    WebTarget target = client.target("http://localhost:8080/api/admin/schedules");

    // Act: POST without credentials
    Response response = target.request().post(Entity.json(testSchedule));

    // Assert: Should get 403 Forbidden
    assertEquals(403, response.getStatus(), "Unauthenticated request should be rejected");
}

@Test
@WithMockUser(roles = "ADMIN")
public void testAdminEndpointAllowsAuthenticatedAdmin() {
    // Arrange: Create request WITH authentication
    WebTarget target = client.target("http://localhost:8080/api/admin/schedules");

    // Act: POST with credentials
    Response response = target.request()
        .header("Authorization", "Bearer " + adminToken)
        .post(Entity.json(testSchedule));

    // Assert: Should succeed
    assertEquals(200, response.getStatus(), "Authenticated admin should be allowed");
}
```

Run tests:
```bash
mvn test -Dtest=*AdminEndpoint* -v
```

#### Step 4: Deploy and verify (30 minutes)

```bash
# Build v1 with fixes
mvn clean package -DskipTests

# Deploy to staging first
kubectl apply -f k8s/staging/schedcu-v1.yaml

# Run smoke tests
curl -v http://staging-schedcu.hospital.local/api/admin/schedules \
  -H "Authorization: Bearer invalid-token"
# Should return 401 Unauthorized

# If OK, deploy to production
kubectl apply -f k8s/production/schedcu-v1.yaml

# Verify production
curl -v http://schedcu.hospital.local/api/health
# Should return 200 OK
```

---

## Security Fix 2: Move Hardcoded Credentials to Vault/Environment Variables

**Duration**: 2 hours
**Risk Level**: Critical
**Impact**: Database, Redis, and API credentials are stored in source code

### Problem Statement

v1 has hardcoded credentials in:
- `application.properties` (database password)
- `environment.xml` (API tokens)
- `RedisConfig.java` (Redis password)

Example (current vulnerable code):
```properties
# application.properties
quarkus.datasource.password=RadiologyDB2024!  # ⚠️ EXPOSED IN SOURCE CODE

# Redis configuration
redis.password=redisSecure123  # ⚠️ EXPOSED

# External API tokens
amion.api.key=sk_live_51234567890  # ⚠️ EXPOSED
```

This exposes credentials to:
- Git history (permanent exposure if committed)
- Build artifacts
- Anyone with source code access
- Log files if accidentally printed

### Solution

Use HashiCorp Vault (or environment variables if Vault unavailable) to manage secrets.

#### Step 1: Audit current credential locations (20 minutes)

```bash
# Find all hardcoded passwords and tokens
grep -r "password\|secret\|token\|key" \
  --include="*.properties" \
  --include="*.xml" \
  --include="*.env" \
  --include="*.java" \
  . | grep -v ".git" | tee /tmp/credentials_audit.txt

# Review results - look for patterns like:
# - "password="
# - "secret="
# - "sk_live_" or "sk_test_" (API keys)
# - Database URLs with embedded passwords
```

Document each finding:
- File path
- Variable name
- Current location (properties/xml/code)
- Sensitivity level (critical/high/medium)

#### Step 2: Set up Vault integration (40 minutes)

**Option A: Using HashiCorp Vault (Recommended)**

1. Create Vault secrets:
```bash
vault kv put secret/schedcu-v1/database \
  url="jdbc:postgresql://db.hospital.local:5432/radiology" \
  username="schedcu_app" \
  password="$(openssl rand -base64 32)"

vault kv put secret/schedcu-v1/redis \
  host="redis.hospital.local" \
  port="6379" \
  password="$(openssl rand -base64 32)"

vault kv put secret/schedcu-v1/amion \
  api_key="sk_live_..." \
  api_secret="..."
```

2. Add Vault dependency to `pom.xml`:
```xml
<dependency>
    <groupId>io.quarkus</groupId>
    <artifactId>quarkus-vault</artifactId>
    <version>3.0.0</version>
</dependency>
```

3. Configure Vault in `application.properties`:
```properties
# Vault configuration
quarkus.vault.url=https://vault.hospital.local:8200
quarkus.vault.auth-method=kubernetes
quarkus.vault.kubernetes-role=schedcu-v1-reader

# Reference secrets
quarkus.datasource.password=${vault:secret/schedcu-v1/database#password}
quarkus.datasource.username=${vault:secret/schedcu-v1/database#username}
quarkus.redis.hosts=${vault:secret/schedcu-v1/redis#host}:${vault:secret/schedcu-v1/redis#port}
```

4. Update Java code to read from Vault:
```java
@Configuration
public class SecretConfig {

    @Inject
    @VaultSecret(path = "secret/schedcu-v1/amion", key = "api_key")
    String amionApiKey;

    @Inject
    @VaultSecret(path = "secret/schedcu-v1/amion", key = "api_secret")
    String amionApiSecret;

    // Use amionApiKey and amionApiSecret in services
}
```

**Option B: Using Environment Variables (if Vault unavailable)**

1. Remove hardcoded values:
```properties
# application.properties (BEFORE)
quarkus.datasource.password=RadiologyDB2024!

# application.properties (AFTER)
quarkus.datasource.password=${DB_PASSWORD}
```

2. Update startup scripts to provide env vars:
```bash
#!/bin/bash
# /etc/systemd/system/schedcu-v1.service

[Service]
Environment="DB_PASSWORD=$(aws secretsmanager get-secret-value --secret-id schedcu-db-password)"
Environment="REDIS_PASSWORD=$(aws secretsmanager get-secret-value --secret-id schedcu-redis-password)"
Environment="AMION_API_KEY=$(aws secretsmanager get-secret-value --secret-id amion-api-key)"

ExecStart=/opt/schedcu/bin/schedcu-v1
```

#### Step 3: Remove hardcoded credentials from source (30 minutes)

1. Clean all properties files:
```bash
# Remove sensitive data
sed -i.bak 's/password=.*/password=${DB_PASSWORD}/g' application.properties
sed -i.bak 's/token=.*/token=${API_TOKEN}/g' environment.xml
sed -i.bak 's/api_key=.*/api_key=${AMION_API_KEY}/g' RedisConfig.java
```

2. Clean Git history (if credentials were committed):
```bash
# CRITICAL: Remove all trace from git history
git filter-repo --paths application.properties --path environment.xml --invert-paths

# Or use BFG Repo Cleaner:
brew install bfg  # or apt install bfg
bfg --delete-files application.properties
bfg --textblob-conversion '.*password.*' '***SECRET***'

# Force push (WARNING: requires force push - coordinate with team)
git push --force-with-lease origin main
```

3. Update `.gitignore` to prevent future commits:
```bash
# .gitignore
.env
application.properties
environment.xml
**/credentials.json
**/secrets/*

# Git hook to prevent accidental secret commits
cat > .git/hooks/pre-commit <<'EOF'
#!/bin/bash
git diff --cached | grep -E "password|api_key|secret|token" && {
  echo "ERROR: Potential secrets found in staged changes"
  echo "Please use environment variables instead"
  exit 1
}
EOF

chmod +x .git/hooks/pre-commit
```

#### Step 4: Test credential loading (30 minutes)

1. Test with Vault:
```bash
# Deploy to staging with Vault enabled
kubectl set env deployment/schedcu-v1 \
  VAULT_ADDR=https://vault.hospital.local:8200 \
  VAULT_TOKEN=$(cat /tmp/vault-token) \
  --record

# Check logs for successful Vault auth
kubectl logs -f deployment/schedcu-v1 | grep -i vault

# Expected: "Vault authentication successful"
```

2. Verify application starts and connects to database:
```bash
# Check health endpoint
curl http://staging-schedcu.hospital.local/api/health
# Should return:
# {
#   "status": "UP",
#   "checks": {
#     "database": "UP",
#     "redis": "UP"
#   }
# }
```

3. Verify no credentials in logs or config:
```bash
kubectl logs deployment/schedcu-v1 | grep -i "password\|secret\|token"
# Should return nothing (no secrets in output)

# Check environment
kubectl exec deployment/schedcu-v1 -- env | grep -E "password|api_key"
# Should return nothing (secrets not exposed)
```

---

## Security Fix 3: Add File Upload Validation (XXE Protection)

**Duration**: 2 hours
**Risk Level**: High
**Impact**: XML/ODS file uploads could allow XML External Entity (XXE) attacks

### Problem Statement

v1 accepts file uploads (ODS schedules) without proper validation:
- No file type validation
- No XXE protection
- No file size limits
- No malware scanning

Example (current vulnerable code):
```java
@Path("/uploads/schedules")
@POST
@Consumes(MediaType.MULTIPART_FORM_DATA)
public Response uploadSchedule(@FormParam("file") InputStream file) {
    // ⚠️ PROBLEM: No validation, XXE vulnerable
    DocumentBuilderFactory factory = DocumentBuilderFactory.newInstance();
    DocumentBuilder builder = factory.newDocumentBuilder();
    Document doc = builder.parse(file);  // XXE VULNERABILITY

    return scheduleService.processSchedule(doc);
}
```

Attacker could craft malicious XML:
```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<!DOCTYPE foo [
  <!ELEMENT foo ANY >
  <!ENTITY xxe SYSTEM "file:///etc/passwd" >]>
<foo>&xxe;</foo>
```

This reads `/etc/passwd` and could expose sensitive files.

### Solution

Implement file upload validation and XXE protection.

#### Step 1: Add file upload handler with validation (40 minutes)

Create `FileUploadValidator.java`:
```java
import java.io.File;
import java.util.*;

public class FileUploadValidator {

    // Allowed extensions for schedule uploads
    private static final Set<String> ALLOWED_EXTENSIONS =
        Set.of("ods", "xls", "xlsx", "csv");

    // Maximum file size: 10 MB
    private static final long MAX_FILE_SIZE = 10 * 1024 * 1024;

    // Magic bytes for file type detection
    private static final Map<String, byte[]> MAGIC_BYTES = Map.ofEntries(
        Map.entry("ods", new byte[]{0x50, 0x4B, 0x03, 0x04}),  // PK.. (ZIP)
        Map.entry("xlsx", new byte[]{0x50, 0x4B, 0x03, 0x04}), // PK.. (ZIP)
        Map.entry("xls", new byte[]{0xD0, 0xCF, 0x11, 0xE0})   // OLE
    );

    public static class ValidationResult {
        public boolean isValid;
        public String errorMessage;
        public String safeFilename;
    }

    public static ValidationResult validate(
        InputStream fileStream,
        String originalFilename,
        String contentType) throws IOException {

        ValidationResult result = new ValidationResult();

        // Step 1: Validate extension
        String extension = getFileExtension(originalFilename).toLowerCase();
        if (!ALLOWED_EXTENSIONS.contains(extension)) {
            result.isValid = false;
            result.errorMessage = String.format(
                "File type not allowed: .%s (allowed: %s)",
                extension,
                String.join(", ", ALLOWED_EXTENSIONS)
            );
            return result;
        }

        // Step 2: Validate content type
        if (!isAllowedContentType(contentType, extension)) {
            result.isValid = false;
            result.errorMessage = "Content type mismatch with file extension";
            return result;
        }

        // Step 3: Validate magic bytes (file signature)
        byte[] magicBytes = readMagicBytes(fileStream);
        if (!verifyMagicBytes(magicBytes, extension)) {
            result.isValid = false;
            result.errorMessage = String.format(
                "File signature does not match .%s format",
                extension
            );
            return result;
        }

        // Step 4: Validate file size
        fileStream.reset();  // Reset stream after magic bytes read
        long fileSize = getStreamSize(fileStream);

        if (fileSize > MAX_FILE_SIZE) {
            result.isValid = false;
            result.errorMessage = String.format(
                "File too large: %d MB (max: %d MB)",
                fileSize / (1024 * 1024),
                MAX_FILE_SIZE / (1024 * 1024)
            );
            return result;
        }

        // Step 5: Generate safe filename (prevent path traversal)
        result.safeFilename = generateSafeFilename(originalFilename);
        result.isValid = true;

        return result;
    }

    private static String getFileExtension(String filename) {
        int lastDot = filename.lastIndexOf('.');
        return lastDot > 0 ? filename.substring(lastDot + 1) : "";
    }

    private static boolean isAllowedContentType(String contentType, String extension) {
        Map<String, List<String>> allowed = Map.ofEntries(
            Map.entry("ods", List.of("application/vnd.oasis.opendocument.spreadsheet")),
            Map.entry("xlsx", List.of("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")),
            Map.entry("xls", List.of("application/vnd.ms-excel")),
            Map.entry("csv", List.of("text/csv", "application/csv"))
        );

        List<String> allowed_types = allowed.getOrDefault(extension, List.of());
        return allowed_types.contains(contentType.toLowerCase());
    }

    private static byte[] readMagicBytes(InputStream stream) throws IOException {
        byte[] magic = new byte[4];
        stream.read(magic);
        return magic;
    }

    private static boolean verifyMagicBytes(byte[] actual, String extension) {
        byte[] expected = MAGIC_BYTES.get(extension);
        if (expected == null) return true;  // No magic bytes check needed

        for (int i = 0; i < Math.min(actual.length, expected.length); i++) {
            if (actual[i] != expected[i]) return false;
        }
        return true;
    }

    private static long getStreamSize(InputStream stream) throws IOException {
        long size = 0;
        byte[] buffer = new byte[1024];
        int bytesRead;
        while ((bytesRead = stream.read(buffer)) != -1) {
            size += bytesRead;
        }
        return size;
    }

    private static String generateSafeFilename(String original) {
        // Remove path traversal attempts
        return original
            .replaceAll("\\.\\.[\\\\/]", "")  // Remove ../
            .replaceAll("[^a-zA-Z0-9._-]", "_")  // Replace special chars
            .substring(0, Math.min(100, original.length()));  // Limit length
    }
}
```

#### Step 2: Disable XXE in XML parsing (20 minutes)

Update the schedule upload endpoint:

```java
@Path("/uploads/schedules")
@POST
@Consumes(MediaType.MULTIPART_FORM_DATA)
public Response uploadSchedule(
    @FormParam("file") InputStream fileStream,
    @FormParam("filename") String filename,
    @HeaderParam("Content-Type") String contentType) throws Exception {

    // Step 1: Validate file
    FileUploadValidator.ValidationResult validation =
        FileUploadValidator.validate(fileStream, filename, contentType);

    if (!validation.isValid) {
        return Response.status(400).entity(Map.of(
            "error", validation.errorMessage
        )).build();
    }

    // Step 2: Configure XXE-safe XML parsing
    DocumentBuilderFactory factory = DocumentBuilderFactory.newInstance();

    // Disable XXE vulnerabilities
    factory.setFeature("http://apache.org/xml/features/disallow-doctype-decl", true);
    factory.setFeature("http://xml.org/sax/features/external-general-entities", false);
    factory.setFeature("http://xml.org/sax/features/external-parameter-entities", false);
    factory.setFeature("http://apache.org/xml/features/nonvalidating/load-external-dtd", false);
    factory.setXIncludeAware(false);
    factory.setExpandEntityReferences(false);

    // Parse with XXE protection enabled
    DocumentBuilder builder = factory.newDocumentBuilder();
    Document doc = builder.parse(fileStream);

    // Step 3: Process with safe filename
    return scheduleService.processSchedule(
        doc,
        validation.safeFilename
    );
}
```

#### Step 3: Create integration tests (40 minutes)

Create `FileUploadSecurityTest.java`:
```java
public class FileUploadSecurityTest {

    @Test
    public void testRejectsInvalidFileExtension() throws Exception {
        InputStream badFile = new ByteArrayInputStream("malware".getBytes());

        FileUploadValidator.ValidationResult result =
            FileUploadValidator.validate(badFile, "evil.exe", "application/octet-stream");

        assertFalse(result.isValid, "Should reject .exe files");
        assertThat(result.errorMessage).contains("not allowed");
    }

    @Test
    public void testRejectsFilesTooLarge() throws Exception {
        // Create 11MB file (exceeds 10MB limit)
        byte[] largeData = new byte[11 * 1024 * 1024];
        InputStream largeFile = new ByteArrayInputStream(largeData);

        FileUploadValidator.ValidationResult result =
            FileUploadValidator.validate(largeFile, "large.ods", "application/vnd.oasis.opendocument.spreadsheet");

        assertFalse(result.isValid, "Should reject files > 10MB");
        assertThat(result.errorMessage).contains("too large");
    }

    @Test
    public void testRejectsXXEAttack() throws Exception {
        // Craft XXE attack payload
        String xxePayload = "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>" +
            "<!DOCTYPE foo [" +
            "  <!ELEMENT foo ANY >" +
            "  <!ENTITY xxe SYSTEM \"file:///etc/passwd\" >" +
            "]>" +
            "<foo>&xxe;</foo>";

        InputStream xmlFile = new ByteArrayInputStream(xxePayload.getBytes());

        // Should parse without resolving external entity
        DocumentBuilderFactory factory = DocumentBuilderFactory.newInstance();
        factory.setFeature("http://xml.org/sax/features/external-general-entities", false);

        DocumentBuilder builder = factory.newDocumentBuilder();
        Document doc = builder.parse(xmlFile);

        // Document should not contain /etc/passwd content
        String content = doc.getDocumentElement().getTextContent();
        assertThat(content).doesNotContain("root");
    }

    @Test
    public void testAcceptsValidODSFile() throws Exception {
        byte[] validODS = createValidODSBytes();
        InputStream file = new ByteArrayInputStream(validODS);

        FileUploadValidator.ValidationResult result =
            FileUploadValidator.validate(file, "schedule.ods", "application/vnd.oasis.opendocument.spreadsheet");

        assertTrue(result.isValid, "Should accept valid ODS files");
        assertNotNull(result.safeFilename);
    }
}
```

Run tests:
```bash
mvn test -Dtest=FileUploadSecurityTest -v
```

#### Step 4: Deploy and verify (20 minutes)

```bash
# Build with fixes
mvn clean package -DskipTests

# Deploy to staging
kubectl apply -f k8s/staging/schedcu-v1.yaml

# Test file upload validation
curl -v -F "file=@test.exe" http://staging-schedcu.hospital.local/api/uploads/schedules
# Should return 400 Bad Request (file type not allowed)

curl -v -F "file=@valid_schedule.ods" http://staging-schedcu.hospital.local/api/uploads/schedules
# Should return 200 OK

# Verify XXE protection works
curl -v -F "file=@xxe_attack.xml" http://staging-schedcu.hospital.local/api/uploads/schedules
# Should return 400 (invalid file type) or process safely without exposing /etc/passwd

# Deploy to production
kubectl apply -f k8s/production/schedcu-v1.yaml
```

---

## Security Fix 4: Add Security Test Suite

**Duration**: 2 hours
**Risk Level**: Medium
**Impact**: Ensure security fixes remain in place and catch future regressions

### Solution

Create comprehensive security test suite.

#### Step 1: Create BaseSecurityTest class (30 minutes)

Create `BaseSecurityTest.java`:
```java
@ExtendWith(QuarkusTest.class)
public abstract class BaseSecurityTest {

    @Inject
    TestClient client;

    protected static final String ADMIN_TOKEN = System.getenv("TEST_ADMIN_TOKEN");
    protected static final String USER_TOKEN = System.getenv("TEST_USER_TOKEN");

    protected Response requestWithAuth(String path, String method, String token) {
        return client.target(path)
            .request()
            .header("Authorization", "Bearer " + token)
            .method(method);
    }

    protected Response requestWithoutAuth(String path, String method) {
        return client.target(path)
            .request()
            .method(method);
    }

    protected void assertUnauthorized(Response response) {
        assertEquals(401, response.getStatus(), "Should return 401 Unauthorized");
    }

    protected void assertForbidden(Response response) {
        assertEquals(403, response.getStatus(), "Should return 403 Forbidden");
    }

    protected void assertOK(Response response) {
        assertEquals(200, response.getStatus(), "Should return 200 OK");
    }
}
```

#### Step 2: Create endpoint security tests (40 minutes)

Create `AdminEndpointSecurityTest.java`:
```java
public class AdminEndpointSecurityTest extends BaseSecurityTest {

    @Test
    public void testAdminEndpointsRequireAuth() {
        // All admin endpoints should require authentication
        String[] adminPaths = {
            "/api/admin/schedules",
            "/api/admin/users",
            "/api/admin/settings"
        };

        for (String path : adminPaths) {
            Response response = requestWithoutAuth(path, "GET");
            assertUnauthorized(response, "GET " + path + " should require auth");

            response = requestWithoutAuth(path, "POST");
            assertUnauthorized(response, "POST " + path + " should require auth");
        }
    }

    @Test
    public void testAdminEndpointsRequireAdminRole() {
        Response response = requestWithAuth(
            "/api/admin/users",
            "DELETE",
            USER_TOKEN  // Regular user token, not admin
        );

        assertForbidden(response, "Regular users should not access admin endpoints");
    }
}
```

#### Step 3: Create credential validation tests (30 minutes)

Create `CredentialInjectionTest.java`:
```java
public class CredentialInjectionTest extends BaseSecurityTest {

    @Test
    public void testNoHardcodedCredentialsInSource() throws IOException {
        // Scan codebase for hardcoded credentials
        File srcDir = new File("src");
        List<File> javaFiles = findFiles(srcDir, "*.java");

        Pattern credentialPattern = Pattern.compile(
            "password\\s*=\\s*['\"].*['\"]|" +
            "token\\s*=\\s*['\"].*['\"]|" +
            "api_key\\s*=\\s*['\"].*['\"]"
        );

        for (File file : javaFiles) {
            String content = new String(Files.readAllBytes(file.toPath()));
            Matcher matcher = credentialPattern.matcher(content);

            assertFalse(matcher.find(),
                "File " + file.getPath() + " contains hardcoded credentials");
        }
    }

    @Test
    public void testApplicationPropertiesDoesNotContainSecrets() throws IOException {
        File propertiesFile = new File("src/main/resources/application.properties");
        String content = new String(Files.readAllBytes(propertiesFile.toPath()));

        assertThat(content)
            .doesNotContain("password=")
            .doesNotContain("api_key=")
            .doesNotContain("token=");
    }
}
```

#### Step 4: Create file upload security tests (20 minutes)

Already covered in Security Fix 3. Run all tests together:

```bash
mvn test -Dtest=*SecurityTest -v
```

#### Step 5: Run security test suite and deploy (20 minutes)

```bash
# Run all security tests
mvn test -Dtest=*SecurityTest,AdminEndpoint*,Credential*,FileUpload* -v

# Expected output: All tests pass ✓
# If any fail, fix and re-run

# Once all pass:
mvn clean package -DskipTests

# Deploy to production with confidence
kubectl apply -f k8s/production/schedcu-v1.yaml

# Verify all 4 fixes in production
./scripts/verify_security_fixes.sh
```

---

## Verification Checklist

Use this checklist to verify all fixes are in place:

```bash
#!/bin/bash
# /scripts/verify_security_fixes.sh

echo "=== v1 Security Fixes Verification ==="

# Fix 1: @PermitAll removed
echo "[1/4] Checking @PermitAll annotations..."
if grep -r "@PermitAll" --include="*.java" src/; then
    echo "❌ FAILED: @PermitAll annotations still present"
    exit 1
else
    echo "✓ PASSED: No @PermitAll annotations found"
fi

# Fix 2: Credentials in Vault/env vars
echo "[2/4] Checking for hardcoded credentials..."
if grep -r "password=" src/main/resources/ | grep -v "\${" | grep -v "#"; then
    echo "❌ FAILED: Hardcoded passwords found"
    exit 1
else
    echo "✓ PASSED: No hardcoded passwords in config files"
fi

# Fix 3: XXE protection
echo "[3/4] Checking XXE protection..."
if grep -r "disallow-doctype-decl" src/ | grep -q "setFeature"; then
    echo "✓ PASSED: XXE protection is enabled"
else
    echo "❌ FAILED: XXE protection not found"
    exit 1
fi

# Fix 4: Security tests
echo "[4/4] Running security test suite..."
mvn test -Dtest=*SecurityTest -q
if [ $? -eq 0 ]; then
    echo "✓ PASSED: All security tests pass"
else
    echo "❌ FAILED: Security tests failed"
    exit 1
fi

echo ""
echo "=== ✓ ALL SECURITY FIXES VERIFIED ==="
```

---

## Deployment Steps

### Week 1 Timeline

**Monday**:
- 9:00-11:00 AM: Fix 1 (@PermitAll removal)
- 11:00 AM-1:00 PM: Fix 2 (Credentials to Vault)
- Lunch

**Tuesday**:
- 9:00-11:00 AM: Fix 3 (XXE protection)
- 11:00 AM-12:00 PM: Fix 4 (Security tests)
- 1:00-5:00 PM: Integration testing & deployment

### Production Deployment

```bash
# Final build
mvn clean package -DskipTests -DskipITs

# Create backup
kubectl get deployment schedcu-v1 -o yaml > /tmp/schedcu-v1-backup.yaml

# Deploy to production
kubectl apply -f k8s/production/schedcu-v1.yaml

# Monitor logs
kubectl logs -f deployment/schedcu-v1 --tail=100

# Run smoke tests
curl http://schedcu.hospital.local/api/health
curl -H "Authorization: Bearer $ADMIN_TOKEN" http://schedcu.hospital.local/api/admin/schedules
```

---

## Rollback Plan

If issues occur after deployment:

```bash
# Immediate rollback to previous version
kubectl rollout undo deployment/schedcu-v1

# Verify rollback successful
kubectl rollout status deployment/schedcu-v1
kubectl logs -f deployment/schedcu-v1 --tail=50
```

---

## Success Criteria

✅ All 4 security fixes implemented
✅ All security tests pass
✅ No breaking changes to existing functionality
✅ Production deployment successful
✅ Admin endpoints require authentication
✅ No hardcoded credentials in code
✅ File uploads validate properly
✅ No XXE vulnerabilities exploitable

---

## Questions?

Reach out to the API & Security Engineer or review:
- `MASTER_PLAN_v2.md` for overall context
- `TEAM_BRIEFING.md` for schedule and team structure
- `week0-spikes/` for dependency validation results

**Status**: Ready for execution in Week 1
**Owner**: API & Security Engineer (1 person, 6-8 hours)
**Parallel Track**: Yes (concurrent with Phase 0 development)
