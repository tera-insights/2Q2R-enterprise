/// <reference path="../../typings/index.d.ts" />
/// <reference path="../services/Auth.ts" />

module admin {

    export class LoginCtrl {

        login(userid: string, password: string) {

            

        }

        static $inject = ['Auth'];
        constructor(private Auth: Auth){
            
        }

    }

}