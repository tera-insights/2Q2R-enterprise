import angular = require('angular');

/**
 * This service manages the applications and the application servers 
 * 
 * @export
 * @class Apps
 */
export class AppSrvc {

    private resource: any = this.$resource("/admin/apps/:id", { id: '@id' }, {
        'get': { method: 'GET' },
        'update': { method: 'PUT', params: { id: '@id' } }
    });

    public get(): ng.IPromise<any> {
        return this.resource.get().$promise;
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
