### Why this is necessary

This process establishes a secure connection using **OAuth 2.0**, the industry standard for API authorization. Before your application can automate tasks (like creating pins), it must obtain an **Access Token**.

We perform this two-step "handshake" to ensure security:

1.  **User Permission:** The user logs in directly on Pinterest to approve the app, so your code never sees their password.
2.  **Scoped Access:** The generated token acts as a secure "digital key" that limits your app to only the specific actions you requested (like `pins:write`), preventing full account access.

### 1. Generate Authorization Code

Copy the authorization link into your browser (fill it in with your app's information). This will redirect you to the Pinterest login page, where you'll need to provide access to your app. You'll then be redirected to your website, where you'll receive a code (check the URL bar).

<img width="478" height="158" alt="unnamed" src="https://github.com/user-attachments/assets/40b9b348-1cf3-4cc7-93ea-442f0ee8acef" />

* **client_id**: Your App ID.
* **redirect_uri**: The URI you have registered in our platform.
* **scope**: Add the necessary scopes that you are going to use, for example to create pins (`boards:read`, `boards:write`, `pins:read`, `pins:write`).

![unnamed](https://github.com/user-attachments/assets/0b2e05aa-d6e2-4fa1-89ff-cf8101540592)


### 2. Exchange Code for Access Token

Now make this call to the API (fill it with your app's information and the code you obtained in step 1).

* **Authorization**: Base64-encoded string made of `app_ID` and `App_secret_key` (You can find them on the Pinterest development page).
* **code**: Use the code you obtained in the previous step.
* **redirect_uri**: Use the exact same URI that you used in step 1.

<img width="1140" height="184" alt="unnamed (1)" src="https://github.com/user-attachments/assets/2a0268a3-a107-4890-a641-1bd951d4ec1f" />
