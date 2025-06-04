# Make Email Sending Fully Configurable

## Summary
Currently, email subjects, templates, and body content are hardcoded throughout the notification domain. This issue proposes making all email-related settings configurable through the existing configuration system.

## Current State Analysis

### Hardcoded Values Found

1. **Email Subjects**
   - Initial notification: `"Encrypted Message from Password Exchange from %s"` (app/internal/domains/notification/domain/service.go:116)
   - Reminder emails: `"Reminder: You have an unviewed encrypted message (Reminder #%d)"` (app/internal/domains/notification/domain/reminder_service.go:264)

2. **Email Body Template** (app/internal/domains/notification/adapters/secondary/smtp/sender.go:56)
   ```go
   Body: fmt.Sprintf("Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about", 
       req.RecipientName, req.SenderName, passwordExchangeURL, passwordExchangeURL)
   ```

3. **Reminder Message Content** (app/internal/domains/notification/domain/reminder_service.go:266)
   ```
   "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message."
   ```

4. **Default Configuration Values** (app/internal/domains/notification/adapters/secondary/viper/config.go)
   - Email template: `/templates/email_template.html`
   - Server email: `server@password.exchange`
   - Server name: `Password Exchange`
   - URL: `https://password.exchange`

### Current Architecture
- Uses hexagonal architecture with clear separation between domain, ports, and adapters
- Configuration accessed through `ConfigPort` interface
- Email sending handled by `SMTPSender` adapter
- Templates stored in `/app/templates/` directory

## Proposed Solution

### 1. Extend ConfigPort Interface
Add new methods to `app/internal/domains/notification/ports/secondary/config.go`:

```go
type ConfigPort interface {
    // Existing methods...
    
    // Email subject configuration
    GetInitialNotificationSubject() string
    GetReminderNotificationSubject() string
    
    // Email body templates
    GetInitialNotificationBodyTemplate() string
    GetReminderNotificationBodyTemplate() string
    
    // Template paths for different email types
    GetReminderEmailTemplate() string
    
    // Email content configuration
    GetReminderMessageContent() string
}
```

### 2. Add Configuration Structure
Extend the configuration to support email settings:

```yaml
email:
  templates:
    initial: "/templates/email_template.html"
    reminder: "/templates/reminder_email_template.html"
  subjects:
    initial: "Encrypted Message from Password Exchange from %s"
    reminder: "Reminder: You have an unviewed encrypted message (Reminder #%d)"
  body:
    initial: "Hi %s, \n %s used our service at <a href=\"%s\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to %s/about"
    reminder: "Please check your original email for the secure decrypt link. For security reasons, the decrypt link cannot be included in reminder emails. If you cannot find the original email, please contact the sender to resend the message."
  sender:
    email: "server@password.exchange"
    name: "Password Exchange"
  url: "https://password.exchange"
```

### 3. Environment Variable Support
Following the existing pattern with `PASSWORDEXCHANGE_` prefix:
- `PASSWORDEXCHANGE_EMAIL_TEMPLATES_INITIAL`
- `PASSWORDEXCHANGE_EMAIL_TEMPLATES_REMINDER`
- `PASSWORDEXCHANGE_EMAIL_SUBJECTS_INITIAL`
- `PASSWORDEXCHANGE_EMAIL_SUBJECTS_REMINDER`
- `PASSWORDEXCHANGE_EMAIL_BODY_INITIAL`
- `PASSWORDEXCHANGE_EMAIL_BODY_REMINDER`

### 4. Implementation Steps

1. **Update ConfigPort interface** with new methods
2. **Update Viper adapter** to implement new config methods with defaults
3. **Update notification service** to use config for subjects
4. **Update SMTP sender** to use config for body templates
5. **Update reminder service** to use config for reminder content
6. **Add template loading flexibility** to support both file paths and inline templates
7. **Write comprehensive tests** for all configuration scenarios
8. **Update documentation** with configuration examples

### 5. Template Enhancement
Allow templates to be specified either as:
- File paths (existing behavior): `/templates/email_template.html`
- Inline templates: `Hi {{.RecipientName}}, {{.SenderName}} sent you a message...`

### 6. Backward Compatibility
- All new configuration will have defaults matching current hardcoded values
- Existing deployments will continue working without configuration changes
- Configuration can be gradually adopted

## Benefits
1. **Flexibility**: Different deployments can customize email content
2. **Localization**: Support for multiple languages through configuration
3. **Branding**: Easy customization for different brands/deployments
4. **Testing**: Easier to test with different configurations
5. **Maintenance**: Changes don't require code modifications

## Testing Requirements
1. Unit tests for all new ConfigPort methods
2. Integration tests with different configurations
3. Tests verifying backward compatibility
4. Template parsing validation tests
5. Environment variable override tests

## Migration Guide
1. Current deployments work without changes (backward compatible)
2. To customize, add email configuration section to config file
3. Override specific values using environment variables
4. Test email sending with new configuration before production deployment

## Implementation Status ✅ COMPLETED

All requirements have been successfully implemented:

- [x] All hardcoded email values are configurable
- [x] Backward compatibility maintained 
- [x] Environment variable overrides work
- [x] Both file-based and inline templates supported
- [x] Comprehensive test coverage
- [x] Documentation updated with examples
- [x] No breaking changes to existing deployments

## Configuration Examples

### Basic Email Configuration

```yaml
email:
  templates:
    initial: "/templates/email_template.html"
    reminder: "/templates/reminder_email_template.html"
  subjects:
    initial: "Encrypted Message from %s"
    reminder: "Reminder #%d: You have an unviewed encrypted message"
  body:
    initial: "Hi %s, %s sent you a secure message. Visit %s to view it."
    reminder: "Please check your original email for the decrypt link."
  sender:
    email: "noreply@yourcompany.com"
    name: "Your Company Secure Exchange"
  url: "https://secure.yourcompany.com"
```

### Inline Template Configuration

```yaml
email:
  subjects:
    initial: "🔒 Secure Message from %s"
    reminder: "⏰ Reminder %d: Unviewed Secure Message"
  body:
    initial: "Hello %s! 👋 %s has sent you a secure message via our platform. Click here to view: %s"
    reminder: "This is a friendly reminder that you have an unviewed secure message. Please check your original email for the access link."
```

### Multi-language Support Example

```yaml
email:
  subjects:
    initial: "Mensaje Cifrado de %s"
    reminder: "Recordatorio #%d: Mensaje No Leído"
  body:
    initial: "Hola %s, %s te ha enviado un mensaje seguro. Visita %s para verlo."
    reminder: "Por favor revisa tu email original para el enlace de descifrado."
```

### Environment Variable Override Examples

```bash
# Override email subjects
export PASSWORDEXCHANGE_EMAIL_SUBJECTS_INITIAL="🔐 New Secure Message from %s"
export PASSWORDEXCHANGE_EMAIL_SUBJECTS_REMINDER="📬 Reminder %d: You have mail!"

# Override body templates  
export PASSWORDEXCHANGE_EMAIL_BODY_INITIAL="Hey %s! %s sent you something secure. Check it out at %s"
export PASSWORDEXCHANGE_EMAIL_BODY_REMINDER="Don't forget to check your secure message!"

# Override sender information
export PASSWORDEXCHANGE_EMAIL_SENDER_NAME="Acme Corp Security"
export PASSWORDEXCHANGE_EMAIL_SENDER_EMAIL="security@acme.com"
```

### Template Flexibility Examples

#### File-based Templates
```yaml
email:
  templates:
    initial: "/custom/templates/branded_email.html"
    reminder: "/custom/templates/reminder.html"
```

#### Inline Templates  
```yaml
email:
  body:
    initial: "<html><body><h1>Hello {{.RecipientName}}!</h1><p>{{.SenderName}} sent you a message.</p></body></html>"
```

### Testing Your Configuration

After updating your configuration, test email sending:

```bash
# Restart the email service
./app email --config=your-config.yaml

# Check logs for configuration loading
tail -f /var/log/password-exchange/email.log
```

## Rollback Instructions

If you need to revert to default behavior:

1. Remove the `email:` section from your config file
2. Unset any `PASSWORDEXCHANGE_EMAIL_*` environment variables  
3. Restart the email service

The system will automatically use the original hardcoded values.