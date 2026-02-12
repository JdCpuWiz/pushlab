# PushLab iOS Client

The iOS client app for PushLab push notification service.

## Features

- User authentication (login/register)
- Automatic APNs device token registration
- Device tag management
- Notification history
- Keychain-based secure token storage

## Setup

### Prerequisites

- Xcode 15 or later
- iOS 17.0 or later
- Apple Developer Account

### Configuration

1. Open `PushLab.xcodeproj` in Xcode
2. Update `Utils/Constants.swift` with your backend URL:
   ```swift
   static let apiBaseURL = "https://your-server.com"
   ```
3. Configure your Team ID and Bundle ID in Xcode project settings
4. Enable Push Notifications capability
5. Build and run on your device

### APNs Setup

1. Go to [Apple Developer Portal](https://developer.apple.com/account)
2. Create an App ID with Push Notifications enabled
3. Create an APNs Authentication Key (.p8 file)
4. Note your Team ID and Key ID
5. Upload the .p8 key to PushLab backend via API

## Project Structure

```
PushLab/
├── App/
│   ├── PushLabApp.swift       # App entry point
│   └── AppDelegate.swift      # APNs registration
├── Views/
│   ├── LoginView.swift        # Login screen
│   ├── RegisterView.swift     # Registration screen
│   ├── HomeView.swift         # Main tab view
│   ├── NotificationHistoryView.swift
│   └── SettingsView.swift     # Device tags, settings
├── Services/
│   ├── APIService.swift       # Backend API client
│   └── AuthService.swift      # Authentication handling
├── Models/
│   ├── User.swift
│   ├── Device.swift
│   └── Notification.swift
└── Utils/
    ├── KeychainHelper.swift   # Secure token storage
    └── Constants.swift        # App constants
```

## Building for Production

1. Update Constants.swift with production backend URL
2. Archive the app in Xcode
3. Upload to App Store Connect
4. Submit for TestFlight or App Store review

## Testing

### Sandbox Environment

For development builds, the app automatically uses the APNs sandbox environment. Make sure your backend has sandbox APNs credentials configured.

### Production Environment

Production builds use the production APNs environment. You'll need separate production APNs credentials uploaded to the backend.

## Troubleshooting

### Not Receiving Notifications

1. Verify notification permissions are granted
2. Check that device token is registered with backend
3. Ensure APNs credentials are uploaded to backend
4. Check backend logs for delivery errors
5. Verify device tags match notification tags

### Device Token Not Registering

1. Verify you're running on a physical device (not simulator)
2. Check network connectivity to backend
3. Verify backend URL is correct in Constants.swift
4. Check console logs for error messages

## License

MIT License
