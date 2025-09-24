# Privacy Policy Implementation Guide

This guide provides options for integrating the privacy policy into your FitTrack application.

## Implementation Options

### Option 1: React Page Component (Recommended)

Create a dedicated privacy policy page in your React app:

```bash
# Create the privacy policy page
touch client/src/routes/privacy.tsx
```

**Benefits:**
- Integrated with your app's navigation
- Consistent styling with your app
- Easy to link from registration/login flows

### Option 2: Static Documentation

Keep the privacy policy as markdown in the docs folder and link to it externally.

**Benefits:**
- Easy to update without app deployment
- Can be hosted separately for legal compliance
- Version control friendly

### Option 3: Modal/Dialog Component

Display the privacy policy in a modal within your app:

**Benefits:**
- Contextual display during registration
- No navigation away from current flow
- Good for consent collection

## Required Updates Before Publishing

### 1. Customize Placeholder Information

Update these sections in `docs/privacy-policy.md`:

```markdown
- [DATE - Please update before publishing] → Actual dates
- [YOUR COMPANY NAME] → Your business name
- [YOUR CONTACT EMAIL] → Your support email
- [YOUR BUSINESS ADDRESS] → Your business address
- [SPECIFY RETENTION PERIOD] → Your data retention period
```

### 2. Legal Review

**Strongly recommended:**
- Have a lawyer review the privacy policy
- Ensure compliance with your target jurisdictions
- Verify third-party service agreements align with disclosures

### 3. Stack Auth Integration

Check Stack Auth's current privacy policy and terms:
- Verify the linked privacy policy URL is current
- Ensure their data practices align with your disclosures
- Update the section if they change their policies

## User Consent Collection

### During Registration

Add privacy policy acceptance to your registration flow:

```typescript
// Example checkbox component
<label className="flex items-center space-x-2">
  <input
    type="checkbox"
    required
    checked={acceptedPrivacy}
    onChange={(e) => setAcceptedPrivacy(e.target.checked)}
  />
  <span className="text-sm">
    I agree to the{' '}
    <Link to="/privacy" className="text-blue-600 hover:underline">
      Privacy Policy
    </Link>
  </span>
</label>
```

### Footer Link

Add privacy policy link to your app footer:

```typescript
// Example footer component
<footer className="text-center text-sm text-gray-600">
  <Link to="/privacy" className="hover:underline">
    Privacy Policy
  </Link>
  {' • '}
  <Link to="/terms" className="hover:underline">
    Terms of Service
  </Link>
</footer>
```

## User Data Deletion

### Implementing User Data Deletion

Your current database schema supports cascade deletion:

```sql
-- Existing CASCADE deletes in schema.sql
REFERENCES users(user_id) ON DELETE CASCADE
```

To implement user account deletion:

1. **API Endpoint:** Create a delete account endpoint
2. **User Interface:** Add account deletion option in settings
3. **Confirmation:** Require confirmation before deletion
4. **Data Export:** Offer data export before deletion

### Example Account Deletion Flow

```typescript
// Example delete account function
const deleteAccount = async () => {
  if (confirm('This will permanently delete all your data. Are you sure?')) {
    try {
      await apiClient.deleteAccount();
      // Logout and redirect
      logout();
      navigate('/');
    } catch (error) {
      console.error('Failed to delete account:', error);
    }
  }
};
```

## Compliance Considerations

### GDPR Compliance (EU Users)

If serving EU users, ensure:
- ✅ Privacy policy covers all GDPR requirements
- ✅ User consent is explicit and documented
- ✅ Data export functionality available
- ✅ Data deletion requests honored within 30 days
- ✅ Data breach notification procedures in place

### CCPA Compliance (California Users)

If serving California users, ensure:
- ✅ Privacy policy includes required CCPA disclosures
- ✅ "Do Not Sell" option available (if applicable)
- ✅ Data access and deletion requests honored
- ✅ Non-discrimination policy in place

## Monitoring and Updates

### Regular Reviews

Schedule regular privacy policy reviews:
- **Quarterly:** Review third-party service changes
- **Annually:** Complete policy audit
- **As Needed:** When adding new features or data collection

### User Notification

When updating the privacy policy:
1. Update the "Last Updated" date
2. Consider email notification for material changes
3. Log the change in your version control
4. Consider showing in-app notification

## Testing Checklist

Before going live:

- [ ] Privacy policy displays correctly on all devices
- [ ] All placeholder text has been replaced
- [ ] Links to third-party policies work
- [ ] Contact information is accurate
- [ ] Legal review completed (recommended)
- [ ] User consent flow tested
- [ ] Data deletion process tested

## Additional Resources

### Privacy Policy Generators

While this policy is custom-built for FitTrack, you may also reference:
- [Termly Privacy Policy Generator](https://termly.io/)
- [PrivacyPolicies.com](https://privacypolicies.com/)

### Legal Resources

- **GDPR:** [Official GDPR Text](https://gdpr-info.eu/)
- **CCPA:** [California Attorney General CCPA Resources](https://oag.ca.gov/privacy/ccpa)
- **Stack Auth Compliance:** Check their documentation for compliance features

---

*Remember: This privacy policy template is a starting point. Always consult with legal counsel for compliance with applicable laws in your jurisdiction.*