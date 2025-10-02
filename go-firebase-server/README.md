# SpendingManagement Go Backend (Firebase Auth)

This folder contains a Gin-based API that trusts Firebase Authentication (Google Sign-In from Flutter).

## Run

```bash
cd go-firebase-server
go mod tidy

# set env
set FIREBASE_PROJECT_ID=<your-project-id>
set GOOGLE_APPLICATION_CREDENTIALS=C:\path	o\serviceAccountKey.json

go run .
# or build
go build -o server.exe
```

## Endpoints

- `GET /healthz`
- `GET /api/me`  (requires Authorization: Bearer <Firebase ID token>)
- `GET /api/spendings` (per-user)
- `POST /api/spendings`  JSON: {amount, category, note?, date?, image_url?, currency?}
- `PUT /api/spendings/:id`
- `DELETE /api/spendings/:id`

## Flutter client snippet

```dart
final googleUser = await GoogleSignIn().signIn();
final googleAuth = await googleUser?.authentication;
if (googleAuth == null) throw 'Canceled Google sign-in';

final credential = GoogleAuthProvider.credential(
  accessToken: googleAuth.accessToken, idToken: googleAuth.idToken,
);
await FirebaseAuth.instance.signInWithCredential(credential);

// Send Firebase ID token to Go backend
final idToken = await FirebaseAuth.instance.currentUser!.getIdToken();
final resp = await http.get(
  Uri.parse('http://localhost:8080/api/me'),
  headers: {'Authorization': 'Bearer $idToken'},
);
```

