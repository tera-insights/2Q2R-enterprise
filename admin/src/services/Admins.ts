/**
  * Interface for admin items
  */
export interface IAdminItem extends ng.resource.IResource<IAdminItem> {
    adminID: string; // The admin ID
    name: string; // the admin's name
    email: string; // the admin's email
    permissions: string[]; // permissions
    role: string; // the admin role, may be "super" (all permissions)
    IV: string; // initialization vector
    seed: string; // for more random password hash
    publicKey: ArrayBuffer; // the public key for 2FA
    $update?: Function; // just so the compiler leaves us alone 
}

export interface IAdminResource extends ng.resource.IResourceClass<IAdminItem> {
    update(params: Object, data: IAdminItem, success?: Function, error?: Function): IAdminItem;
}

/**
 * This service manages admins (if the current admin has permissions to do so) 
 * 
 * @export
 * @class Admins
 */
export class Admins {

    public resource: IAdminResource; // the resource to access backend

    static Resource($resource: ng.resource.IResourceService): IAdminResource {
        var resource = $resource("/admin/admins/:id", { id: '@id' }, {
            'update': { method: 'PUT', params: { id: '@id' } }
        });
        return <IAdminResource>resource;
    }

    static $inject = ['$resource', '$q', '$http'];

    constructor(
        private $resource: ng.resource.IResourceService,
        private $q: angular.IQService,
        private $http: angular.IHttpService
    ) {
        this.resource = Admins.Resource($resource);
    }

}