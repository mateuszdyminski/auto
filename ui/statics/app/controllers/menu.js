'use strict';

angular.module('FC').controller('MenuCtrl', function($scope, $location) {
    $scope.isNavCollapsed = true;
    $scope.isCollapsed = false;
    $scope.isCollapsedHorizontal = false;

    $scope.isActive = function(viewLocation) {
        return viewLocation === $location.path();
    };
});