import 'angular-resource';
import { createAuthenticator, createSigner, ExternalKeyPair } from 'p256-auth';

/**************
 * Interfaces *
 **************/

interface IServerKey {
    public: string;
    expires: number;
}

interface INonce {
    nonce: string;
    expires: number;
}

/**
 * Service responsible for authentication and maintaining sessions. 
 * 
 * @author Sam Claus
 * @version 1/18/17
 * @copyright Tera Insights, LLC
 */
export class AuthSrvc {

    private resource: any = this.$resource('', { headers: this.authHeaders }, {
        public: { method: 'GET', url: '/admin/public' },
        nonce: { method: 'GET', url: '/admin/nonce/:id' }
    });

    private ephemeralKeyManager = createAuthenticator();
    private signingKeyManager = createSigner();

    private authHeaders: any;
    private adminID: string;

    /**
     * This method prepares first-factor authentication headers for
     * second factor authentication requests and should be called once
     * the admin submits their signing key and enters their password.
     * @param {ExternalKeyPair} key      Signing key from file.
     * @param {Uint8Array}      password Password from secure field;
     *                                   used to unwrap @key.
     * @returns A promise that fulfills when first factor headers have
     *          been prepared and are available globally from this service.
     * @throws If the user enters an incorrect password or their key file
     *         is corrupted or incorrect altogether. The error message will
     *         explain which error occurred and should be displayed in a toast.
     */
    public prepareFirstFactor(key: ExternalKeyPair, password: Uint8Array): ng.IPromise<void> {
        let deferred = this.$q.defer<void>();
        this.adminID = ''; // TODO: what is an adminID?

        this.$q.all([
            this.getServerPublic(),
            this.getNonce()
        ]).then(([serverPublic, nonce]) => {
            Promise.all([
                this.ephemeralKeyManager.generateKeyPair(),
                this.ephemeralKeyManager.importServerKey(serverPublic.public),
                this.signingKeyManager.importKey(key, password)
            ]).then(() => {
                this.ephemeralKeyManager.getPublic().then(ephemeralPublic => {
                    this.signingKeyManager.sign(ephemeralPublic, 'base64URL').then(signature => {
                        this.authHeaders = {
                            'X-Authentication-Type': 'admin-frontend',
                            'X-Public-Key': ephemeralPublic,
                            'X-Public-Signature': signature
                        };
                        deferred.resolve();
                    });
                });
            });
        });

        return deferred.promise;
    }

    /**
     * Because an HMAC signature must be computed for every message,
     * the authentication headers will change with every request to the
     * server. This method modifies the headers each time based upon the
     * given message.
     * @param {any} message Request needing authentication.
     * @returns A promise fulfilling with a headers object.
     */
    public getAuthHeaders(message: any): ng.IPromise<any> {
        let deferred = this.$q.defer<any>();
        let msgString = JSON.stringify(message);

        this.getNonce().then(nonce => {
            this.ephemeralKeyManager.computeHMAC(msgString).then(hmac => {
                this.authHeaders['X-Nonce'] = nonce.nonce;
                this.authHeaders['X-authentication'] = this.adminID + ':' + hmac;
                deferred.resolve(this.authHeaders);
            });
        });

        return deferred.promise;
    }

    private getServerPublic(): ng.IPromise<IServerKey> {
        return this.resource.public().$promise;
    }

    private getNonce(): ng.IPromise<INonce> {
        return this.resource.nonce({ id: this.adminID }).$promise;
    }

    static $inject = [
        '$resource',
        '$q'
    ];

    constructor(
        private $resource: ng.resource.IResourceService,
        private $q: ng.IQService
    ) { }
}

