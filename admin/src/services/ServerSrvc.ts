/**************
 * Interfaces *
 **************/

export interface INewServerRequest {
    appID: string;
    baseURL: string;
    keyType: string;
    publicKey: string;
    permissions: string;
}

export interface IServerUpdateRequest {
    serverID: string;
    baseURL: string;
    keyType: string;
    publicKey: string;
    permissions: string;
}

export interface IServerInfo {
    serverID: string;
    appID: string;
    baseURL: string;
    keyType: string;
    publicKey: string;
    permissions: string;
}

/**
 * This service manages physical application servers.
 * 
 * @author Sam Claus
 * @version 1/18/17
 * @copyright Tera Insights, LLC
 */
export class ServerSrvc {
    private resource: any = this.$resource("/admin/servers/:id", {}, {
        'query':  { method: 'GET',    url: '/admin/server', isArray: true },
        'create': { method: 'POST',   url: '/admin/server' },
        'update': { method: 'PUT',    url: '/admin/server/:id' },
        'delete': { method: 'DELETE', url: '/admin/server/:id' }
    });

    public query(): ng.IPromise<IServerInfo[]> {
        return this.resource.query().$promise;
    }

    public create(server: INewServerRequest): ng.IPromise<IServerInfo> {
        return this.resource.create({}, server).$promise;
    }

    public update(server: IServerUpdateRequest): ng.IPromise<IServerInfo> {
        return this.resource.update({ id: '@serverID' }, server).$promise;
    }

    public delete(serverID: string): ng.IPromise<string> {
        return this.resource.delete({ id: serverID }).$promise;
    }

    static $inject = [
        '$resource',
        '$q'
    ];

    constructor(
        private $resource: ng.resource.IResourceService,
        private $q: angular.IQService
    ) {}

}