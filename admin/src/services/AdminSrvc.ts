import 'angular-resource';

/**************
 * Interfaces *
 **************/

export interface INewAdminRequest {
    name: string;
    email: string;
    permissions: string[];
    adminFor: string;
    iv: string;
    salt: string;
    publicKey: string;
    signingPublicKey: string;
    signature: string;
}

export interface INewAdminReply {
    requestID: string;
}

export interface IAdminUpdateRequest {
    adminID: string;
    name: string;
    email: string;
    primarySigningKeyID: string;
    adminFor: string;
}

export interface IAdminRoleChangeRequest {
    adminID: string;
    role: string;
    status: string;
    permissions: string;
    adminFor: string;
}

export interface IAdminDeleteReply {
    numAffected: number;
}

export interface IAdminInfo {
    activeID: string;
    status: 'active' | 'inactive';
    name: string;
    email: string;
    permissions: string;
    role: string;
    primarySigningKeyID: string;
    adminFor: string;
}

/**
 * This service manages admins, with some functions
 * limited to super-admin use.
 * 
 * @author Sam Claus
 * @author Alin Dobra
 * @version 1/18/17
 * @copyright Tera Insights, LLC
 */
export class AdminSrvc {

    private resource: any = this.$resource('', {}, {
        'query':  { method: 'GET',    url: '/admin/admin', isArray: true },
        'create': { method: 'POST',   url: '/admin/new' },
        'roles':  { method: 'POST',   url: 'admin/admin/roles' },
        'update': { method: 'PUT',    url: '/admin/admin/:id' },
        'delete': { method: 'DELETE', url: '/admin/admin/:id' }
    });

    public query(): ng.IPromise<IAdminInfo> {
        return this.resource.query().$promise;
    }

    public create(req: INewAdminRequest): ng.IPromise<INewAdminReply> {
        return this.resource.create({}, req).$promise;
    }

    public changeRoles(req: IAdminRoleChangeRequest): ng.IPromise<IAdminInfo> {
        return this.resource.roles({}, req).$promise;
    }

    public update(req: IAdminUpdateRequest): ng.IPromise<IAdminInfo> {
        return this.resource.update({ id: '@adminID' }, req).$promise;
    }

    public delete(adminID: string): ng.IPromise<IAdminDeleteReply> {
        return this.resource.delete({ id: adminID }).$promise;
    }

    static $inject = [
        '$resource',
        '$q'
    ];

    constructor(
        private $resource: ng.resource.IResourceService,
        private $q: ng.IQService
    ) {}

}