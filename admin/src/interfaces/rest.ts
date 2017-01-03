/**
 * This file contains interfaces with the backend. This file mimics
 * io_interfaces.go
 */

module admin {

    // NewAppRequest is the request to `POST /v1/app/new`.
    export interface INewAppRequest {
        appName: string;
    }

    // NewAppReply is the response to `POST /v1/app/new`.
    export interface INewAppReply {
        appID: string;
    }

    // AppIDInfoReply is the reply to `GET /v1/info/:appID`.
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

    // NewServerRequest is the request to `POST /v1/admin/server/new`.
    export interface INewServerRequest {
        serverName: string;
        appID: string;
        baseURL: string;
        keyType: string;
        publicKey: string; // base-64 encoded byte array
        permissions: string;
    }

    // NewServerReply is the response to `POST `/v1/admin/server/new`.
    export interface INewServerReply {
        serverName: string;
        serverID: string;
    }

    // DeleteServerRequest is the request to `POST /v1/admin/server/delete`.
    export interface IDeleteServerRequest {
        serverID: string;
    }

    // AppServerInfoRequest is the request to `POST /v1/admin/server/info`.
    export interface IAppServerInfoRequest {
        serverID: string;
    }

    // RegistrationSetupReply is the reply to `GET /v1/register/request/:userID`.
    export interface RegistrationSetupReply {
        // base64Web encoded random reply id
        id: string;

        // Url at which the registration iframe can be found. Pass to frontend.
        registerUrl: string;
    }

    // RegisterRequest is the request to `POST /v1/register`.
    export interface IRegisterRequest {
        successful: boolean;
        // Either a successfulRegistrationData or a failedRegistrationData
        data: ISuccessfulRegistrationData | IFailedRegistrationData;
    }

    // RegisterResponse is the response to `POST /v1/register`.
    export interface IRegisterResponse {
        successful: boolean;
        message: string;
    }

    export interface ISuccessfulRegistrationData {
        clientData: string;     // base64 serialized client data
        registrationData: string; // base64 binary registration data
        deviceName: string;
        type: string;     // device export interface and key type
        fcmToken: string; // Firebase Communicator Device token
    }

    export interface IFailedRegistrationData {
        errorMessage: string;
        errorStatus: number;
    }

    // NewUserRequest is the request to `POST /v1/admin/user/new`.
    export interface NewUserRequest {
    }

    // NewUserReply is the reply to `POST /v1/admin/user/new`.
    export interface INewUserReply {
        userID: string;
    }

    // AuthenticationSetupRequest is the request to `POST /v1/auth/request`.
    export interface IAuthenticationSetupRequest {
        appID: string;
        timestamp: Date;
        userID: string;
        keyID: string;
        authentication: any; // TODO: was AuthenticationData
    }

    // AuthenticationSetupReply is the response to `POST /v1/auth/request`.
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

    // Reply to `GET /v1/users/:userID`
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
}