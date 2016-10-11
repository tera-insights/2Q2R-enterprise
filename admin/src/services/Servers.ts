/// <reference path="../../typings/index.d.ts" />

module admin {

    /**
     * Interface for server collection items
     */
    export interface IServerItem extends ng.resource.IResource<IServerItem> {
        serverName: string; // displayable name
        appID: string; // the appID of the application the server is associated with
        baseURL: string; // the server domain
        keyType: string; // what key type is being used for Diffie-Hellman with the server
        publicKey: string; // generated in browser for Diffie-Hellman
        permissions: string; // server permissions
        $update?: Function; // just so the compiler leaves us alone 
    }

    export interface IServerResource extends ng.resource.IResourceClass<IServerItem> {
        update(params: Object, data: IServerItem, success?: Function, error?: Function): IServerItem;
    }

    /**
     * This service manages physical application servers
     * 
     * @export
     * @class Servers
     */
    export class Servers {
        public resource: IServerResource; // the resource to access backend

        static Resource($resource: ng.resource.IResourceService): IServerResource {
            var resource = $resource("/admin/servers/:id", { id: '@id' }, {
                'update': { method: 'PUT', params: { id: '@id' } }
            });
            return <IServerResource>resource;
        }

        static $inject = ['$resource', '$q', '$http'];

        constructor($resource: ng.resource.IResourceService,
            private $q: angular.IQService,
            private $http: angular.IHttpService) {
            this.resource = admin.Servers.Resource($resource);
        }

    }
}