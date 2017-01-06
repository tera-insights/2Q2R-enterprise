//////////////////////////////////////////
//
//    Usage
//
//    <secure-password 
//        array="var name of TypedArray"
//        placeholder="string"
//        filler="string"
//        idlemsec="number"
//        idlesec="number"
//        enter="functionInParentScope()"
//        change="functionInParentScope()" 
//        extended
//        autofocus >
//    </secure-password>
//
//

angular.module('angular-secure-password', []).directive('securePassword', ['$timeout', '$sce',
  function($timeout, $sce){
    return {  
      restrict: 'E',
      replace: false,
      transclude: false,
      scope: {buffer: '=array', enter: '&', change: "&"},

      template: '<div style="margin-bottom: 10px;">'+
                  '<span class="secure-input" ng-keypress="append($event)" ng-keydown="alter($event)" tabindex="0">'+
                    '<span ng-repeat="char in notPassword">'+
                      '<span ng-if="$first" class="first-symbol">&nbsp;</span>'+
                      '<span class="password-symbol" ng-bind-html="filler">&nbsp;</span>'+
                    '</span>'+
                    '<span class="blinking-cursor">|&nbsp;</span>'+
                    '<span class="placeholder" ng-if="notPassword.length===0">{{placeholder}}</span>'+
                  '</span>'+
                '</div>',
                
      link: {
        pre: function(scope, element, attrs){
          var cursor = 0;
          var size = ArrayBuffer.isView(scope.buffer) ? scope.buffer.length : 0;
          scope.notPassword = [];
          
          scope.placeholder = attrs.placeholder || 'Password';
          var extended = typeof attrs.extended === 'string' || false; //if extended is given, assume true
          
                      // parseInt('nonumberhere') => NaN, NaN*1000 => NaN, NaN || 3 => 3
          var idletime = parseInt(attrs.idlemsec, 10) || parseInt(attrs.idlesec, 10)*1000 || 0;
          var idlePromise = false;
          
          // can't pre/ap-pend &/; in the original directive call because the HTML actually renders the character into its string form
          //    e.g., <secure-password filler="&nbsp;"> is equivalent to <secure-password filler=" ">, so scope.filler becomes &bull; from the |or| statement
          scope.filler = $sce.trustAsHtml('&'+(attrs.filler || 'bull')+';');
          
          // scope.tabindex = parseInt(attrs.tabindex) || 0;  // // doesn't actually affect anything yet
          
          function clear() {
            console.log('cleared!');
            cursor = 0;
            if(size !== 0)
              scope.buffer.fill(0);
            scope.notPassword = [];
            return false;
          }
          
          function resetTimeout(){
            if(idletime === 0) {
              return;
            } else if(typeof idlePromise === 'boolean') {
              idlePromise = $timeout(clear, idletime);
            } else {
              idlePromise = $timeout.cancel(idlePromise);
              resetTimeout();
            }
          }
          
          scope.alter = function(keyEvent){
            if (keyEvent.keyCode === 8 && cursor > 0) { // || keyEvent.keyIdentifier === 'U+0008')
              scope.buffer[--cursor] = 0;
              scope.notPassword.splice(-1, 1);
            }
            if (keyEvent.keyCode === 8 && cursor >= 0)
              stopEvent(keyEvent); // to prevent backspace from navigating backwards in browser history
            if (keyEvent.keyCode === 13 && typeof scope.enter === 'function') {
              scope.enter();
              if(idlePromise)
                idlePromise = $timeout.cancel(idlePromise);
              stopEvent();
            } else
              resetTimeout(); //even though user didn't enter valid key, user did interact with input
            if ( typeof scope.change === 'function')
              scope.change(); // signal the change to the callee
          };
          
          //// NOTE: Functionality of extended ASCII has not been tested **at all**
          scope.append = function(keyEvent) {
            var cC = keyEvent.charCode;
            var cond1 = cC >= 32 && cC <= 126,  //excludes 'DEL' char at 127
                cond2 = cC >= 32 && cC <= 255 && (cC !== 129 && cC !== 141 && cC !== 143 && cC !== 144 && cC !== 157);
            if (cursor >= 0 && cursor < size) {
              if(cond1 || (scope.extended === true && cond2)) {
                scope.buffer[cursor] = cC;
                scope.notPassword.push(cursor);
                cursor++;
                stopEvent(keyEvent);
              }
            }
          };
          
          scope.$watch('buffer[0]', function(newVal){
            if(newVal === 0) {  // .fill(0) called from parent
              scope.notPassword = []; // reset corresponding vars
              cursor = 0;
              if(idlePromise)
                idlePromise = $timeout.cancel(idlePromise);
            }
          });
          
          var stopEvent = function(e) { if (e) { e.preventDefault(); if (e.stopPropagation) { e.stopPropagation(); } else { e.cancelBubble = true; } } };
        },
        
        post: function(scope, element, attrs){
          if(typeof attrs.autofocus === 'string') //if it exists at all (including empty string)
            // NOTE: using querySelector, NOT querySelectorAll, it will choose first one and that's that.
            document.querySelector('secure-password[autofocus] > div > span').focus();
        }
      }
    };
  }
]);