/**
 * Interface for user collection items
 */
export interface IUserItem extends ng.resource.IResource<IUserItem> {
    appName: string; // displayable name
    id: string; // the item ID
    $update?: Function; // just so the compiler leaves us alone 
}

export interface IUserResource extends ng.resource.IResourceClass<IUserItem> {
    update(params: Object, data: IUserItem, success?: Function, error?: Function): IUserItem;
}

/**
 * This service manages the users across all applications
 * 
 * @export
 * @class Users
 */
export class Users {
    public resource: IUserResource; // the resource to access backend

    static Resource($resource: ng.resource.IResourceService): IUserResource {
        var resource = $resource("/admin/app/:id", { id: '@id' }, {
            'update': { method: 'PUT', params: { id: '@id' } }
        });
        return <IUserResource>resource;
    }

    static $inject = ['$resource', '$q', '$http'];

    constructor($resource: ng.resource.IResourceService,
        private $q: angular.IQService,
        private $http: angular.IHttpService) {
        this.resource = Users.Resource($resource);
    }

}
