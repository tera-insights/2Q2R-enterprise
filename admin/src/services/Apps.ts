/// <reference path="../../typings/index.d.ts" />

/**
  * Interface for todo items
  */
export interface IAppItem extends ng.resource.IResource<IAppItem> {
    appName: string; // displayable name
    appID?: string; // The app ID 
    $update?: Function; // just so the compiler leaves us alone 
}

export interface IAppResource extends ng.resource.IResourceClass<IAppItem> {
    update(params: Object, data: IAppItem, success?: Function, error?: Function): IAppItem;
}

/**
 * This service manages the applications and the application servers 
 * 
 * @export
 * @class Apps
 */
export class Apps {

    public resource: IAppResource; // the resource to access backend

    static Resource($resource: ng.resource.IResourceService): IAppResource {
        var resource = $resource("/admin/apps/:id", { id: '@id' }, {
            'update': { method: 'PUT', params: { id: '@id' } }
        });
        return <IAppResource>resource;
    }

    static $inject = ['$resource', '$q', '$http'];

    constructor(
        private $resource: ng.resource.IResourceService,
        private $q: angular.IQService,
        private $http: angular.IHttpService
    ) {
        this.resource = Apps.Resource($resource);
    }

}
