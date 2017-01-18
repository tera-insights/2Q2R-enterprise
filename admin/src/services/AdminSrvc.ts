import { INewAdminRequest, INewAdminReply } from '../interfaces/rest';

import 'angular-resource';

/**
  * Interface for admin items.
  */
export interface IAdminInfo {
    adminID: string; // The admin ID
    name: string; // the admin's name
    email: string; // the admin's email
    permissions: string[]; // permissions
    role: string; // the admin role, may be "super" (all permissions)
    IV: string; // initialization vector
    seed: string; // for more random password hash
    publicKey: ArrayBuffer; // the public key for 2FA
}

/**
 * This service manages admins (if the current admin has permissions to do so).
 * 
 * @author Sam Claus
 * @author Alin Dobra
 * @version 1/17/17
 * @copyright Tera Insights, LLC
 */
export class AdminSrvc {

    private resource: any = this.$resource("/admin/admins/:id", { id: '@id' }, {
        'get': { method: 'GET' },
        'update': { method: 'PUT', params: { id: '@id' } },
        'create': { method: 'POST', url: '/admin/new' }
    });

    public get(id: string): ng.IPromise<IAdminInfo> {
        return this.resource.get({ id: id }).$promise;
    }

    public update(admin: IAdminInfo): ng.IPromise<IAdminInfo> {
        return this.resource.update(admin).$promise;
    }

    public create(request: INewAdminRequest): ng.IPromise<INewAdminReply> {
        return this.resource.create(request).$promise;
    }

    static $inject = [
        '$resource'
    ];

    constructor(
        private $resource: ng.resource.IResourceService
    ) {}

}