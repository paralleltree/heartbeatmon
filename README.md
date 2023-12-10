# heartbeatmon

## Setting up

### Build and Deploy an image

  1. Enter to `cmd/lambda`
  2. Run `sam build` to build container image
  3. Run `sam deploy --guided` to create resources.
  4. Add settings to get heartrate from Fitbit and store

### Put settings into a S3 Bucket

Run `cmd/auth/main.go` to authenticate Fitbit API and write tokens to file below:

  * `clientCredentials.json`
    * `clientId`
    * `clientSecret`
  * `accessToken.json`
    * `accessToken`
    * `refreshToken`

Upload these files to root of the S3 bucket created by previous step.
