/**
 * This file contains interfaces with the backend. This file mimics
 * io_interfaces.go
 */

// Request to `POST /admin/new`
export interface INewAdminRequest {
    name: string;
    email: string;
    permissions?: string[];
    adminFor: string; // appID
    iv: string;
    salt: string;
    publicKey: string;
    signingPublicKey?: string;
    signature?: string;
}

// Response to `POST /admin/new`
export interface INewAdminReply {
	requestID: String;
}

// Request to `POST /v1/app/new`.
export interface INewAppRequest {
    appName: string;
}

// Response to `POST /v1/app/new`.
export interface INewAppReply {
    appID: string;
}

// Response to `GET /v1/info/:appID`.
export interface IAppIDInfoReply {
    // string specifying displayable app name
    appName: string;

    // string specifying the prefix of all routes
    baseURL: string;

    appURL: string;

    appID: string;

    // base 64 encoded public key of the 2Q2R server
    serverPubKey: string;

    // Only P256 supported for now
    serverKeyType: string;
}

// Request to `POST /v1/admin/server/new`.
export interface INewServerRequest {
    serverName: string;
    appID: string;
    baseURL: string;
    keyType: string;
    publicKey: string; // base-64 encoded byte array
    permissions: string;
}

// Response to `POST `/v1/admin/server/new`.
export interface INewServerReply {
    serverName: string;
    serverID: string;
}

// Request to `POST /v1/admin/server/delete`.
export interface IDeleteServerRequest {
    serverID: string;
}

// Request to `POST /v1/admin/server/info`.
export interface IAppServerInfoRequest {
    serverID: string;
}

// Response to `GET /v1/register/request/:userID`.
export interface IRegistrationSetupReply {
    // base64Web encoded random reply id
    id: string;

    // Url at which the registration iframe can be found. Pass to frontend.
    registerUrl: string;
}

// Request to `POST /v1/register`.
export interface IRegisterRequest {
    successful: boolean;
    // Either a successfulRegistrationData or a failedRegistrationData
    data: ISuccessfulRegistrationData | IFailedRegistrationData;
}

// Response to `POST /v1/register`.
export interface IRegisterResponse {
    successful: boolean;
    message: string;
}

export interface ISuccessfulRegistrationData {
    clientData: string;     // base64 serialized client data
    registrationData: string; // base64 binary registration data
    deviceName: string;
    type: string;     // device export interface and key type
    fcmToken: string; // Firebase StatsSrvc Device token
}

export interface IFailedRegistrationData {
    errorMessage: string;
    errorStatus: number;
}

// Request to `POST /v1/admin/user/new`.
export interface NewUserRequest {
}

// Response to `POST /v1/admin/user/new`.
export interface INewUserReply {
    userID: string;
}

// Request to `POST /v1/auth/request`.
export interface IAuthenticationSetupRequest {
    appID: string;
    timestamp: Date;
    userID: string;
    keyID: string;
    authentication: any; // TODO: was AuthenticationData
}

// Response to `POST /v1/auth/request`.
export interface IAuthenticationSetupReply {
    // base64Web encoded random reply id
    id: string;

    // Url at which the registration iframe can be found. Pass to frontend.
    authUrl: string;
}

export interface IAuthenticateRequest {
    successful: boolean;
    data: ISuccessfulAuthenticationData | IFailedAuthenticationData;
}

// Request to `POST /v1/auth/{requestID}/challenge`
export interface ISetKeyRequest {
    keyID: string;
}

// Response to `POST /v1/auth/{requestID}/challenge`
export interface ISetKeyReply {
    keyID: string;
    challenge: string;
    counter: number;
    appID: string;
}

// Response to `GET /v1/users/:userID`
export interface IUserExistsReply {
    exists: boolean;
}

export interface ISuccessfulAuthenticationData {
    clientData: string;
    signatureData: string;
}

export interface IFailedAuthenticationData {
    challenge: string;
    errorMessage: string;
    errorStatus: number;
}