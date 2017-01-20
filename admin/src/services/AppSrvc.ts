import 'angular-resource';

/**************
 * Interfaces *
 **************/

export interface INewAppRequest {
    appName: string;
}

export interface IAppUpdateRequest {
    appID: string;
    appName: string;
}

export interface IAppInfo {
    appID: string;
    appName: string;
}

/**
 * This service manages the applications under the
 * 2Q2R server.
 * 
 * @author Sam Claus
 * @version 1/18/17
 * @copyright Tera Insights, LLC
 */
export class AppSrvc {

    private resource: any = this.$resource('', {}, {
        'query':  { method: 'GET',    url: '/admin/app', isArray: true },
        'create': { method: 'POST',   url: '/admin/app' },
        'update': { method: 'POST',   url: '/admin/app/:id' },
        'delete': { method: 'DELETE', url: '/admin/app/:id' }
    });

    public query(): ng.IPromise<IAppInfo[]> {
        return this.resource.query().$promise;
    }

    public create(req: INewAppRequest): ng.IPromise<IAppInfo> {
        return this.resource.create({}, req).$promise;
    }

    public update(req: IAppUpdateRequest): ng.IPromise<IAppInfo> {
        return this.resource.update({ id: '@appID' }, req).$promise;
    }

    public delete(appID: string): ng.IPromise<string> {
        return this.resource.delete({ id: appID }).$promise;
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
