import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'

import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { BreadcrumbsProps, BreadcrumbSetters } from './components/Breadcrumbs'
import { LayoutProps } from './Layout'
import { ExtensionAlertProps } from './repo/RepoContainer'
import { android } from './repogroups/Android'
import { cncf } from './repogroups/cncf'
import { golang } from './repogroups/Golang'
import { kubernetes } from './repogroups/Kubernetes'
import { python2To3Metadata } from './repogroups/Python2To3'
import { reactHooks } from './repogroups/ReactHooks'
import { RepogroupPage } from './repogroups/RepogroupPage'
import { stackStorm } from './repogroups/StackStorm'
import { stanford } from './repogroups/Stanford'
import { temporal } from './repogroups/Temporal'
import { StreamingSearchResults } from './search/results/streaming/StreamingSearchResults'
import { isMacPlatform, UserExternalServicesOrRepositoriesUpdateProps } from './util'
import { lazyComponent } from './util/lazyComponent'

const SearchPage = lazyComponent(() => import('./search/input/SearchPage'), 'SearchPage')
const SearchResults = lazyComponent(() => import('./search/results/SearchResults'), 'SearchResults')
const SiteAdminArea = lazyComponent(() => import('./site-admin/SiteAdminArea'), 'SiteAdminArea')
const ExtensionsArea = lazyComponent(() => import('./extensions/ExtensionsArea'), 'ExtensionsArea')
const SearchConsolePage = lazyComponent(() => import('./search/SearchConsolePage'), 'SearchConsolePage')
const SignInPage = lazyComponent(() => import('./auth/SignInPage'), 'SignInPage')
const SignUpPage = lazyComponent(() => import('./auth/SignUpPage'), 'SignUpPage')
const PostSignUpPage = lazyComponent(() => import('./auth/PostSignUpPage'), 'PostSignUpPage')
const SiteInitPage = lazyComponent(() => import('./site-admin/init/SiteInitPage'), 'SiteInitPage')

export interface LayoutRouteComponentProps<RouteParameters extends { [K in keyof RouteParameters]?: string }>
    extends RouteComponentProps<RouteParameters>,
        Omit<LayoutProps, 'match'>,
        BreadcrumbsProps,
        BreadcrumbSetters,
        ExtensionAlertProps,
        UserExternalServicesOrRepositoriesUpdateProps {
    isSourcegraphDotCom: boolean
    isRedesignEnabled: boolean
}

export interface LayoutRouteProps<Parameters_ extends { [K in keyof Parameters_]?: string }> {
    path: string
    exact?: boolean
    render: (props: LayoutRouteComponentProps<Parameters_>) => React.ReactNode

    /**
     * A condition function that needs to return true if the route should be rendered
     *
     * @default () => true
     */
    condition?: (props: LayoutRouteComponentProps<Parameters_>) => boolean
}

// Force a hard reload so that we delegate to the serverside HTTP handler for a route.
function passThroughToServer(): React.ReactNode {
    window.location.reload()
    return null
}

/**
 * Holds all top-level routes for the app because both the navbar and the main content area need to
 * switch over matched path.
 *
 * See https://reacttraining.com/react-router/web/example/sidebar
 */
export const routes: readonly LayoutRouteProps<any>[] = [
    {
        path: '/',
        render: props =>
            window.context.sourcegraphDotComMode && !props.authenticatedUser ? (
                <Redirect to="https://about.sourcegraph.com" />
            ) : (
                <Redirect to="/search" />
            ),
        exact: true,
    },
    {
        path: '/search',
        render: props => (props.parsedSearchQuery ? <StreamingSearchResults {...props} /> : <SearchPage {...props} />),
        exact: true,
    },
    {
        path: '/search/query-builder',
        render: props =>
            props.showQueryBuilder ? (
                lazyComponent(() => import('./search/queryBuilder/QueryBuilderPage'), 'QueryBuilderPage')(props)
            ) : (
                <Redirect to="/search" />
            ),
        exact: true,
    },
    {
        path: '/search/console',
        render: props =>
            props.showMultilineSearchConsole ? (
                <SearchConsolePage {...props} isMacPlatform={isMacPlatform} />
            ) : (
                <Redirect to="/search" />
            ),
        exact: true,
    },
    {
        path: '/sign-in',
        render: props => <SignInPage {...props} context={window.context} />,
        exact: true,
    },
    {
        path: '/sign-up',
        render: props => <SignUpPage {...props} context={window.context} />,
        exact: true,
    },
    {
        path: '/post-sign-up',
        render: props => <PostSignUpPage {...props} context={window.context} />,
        exact: true,
    },
    {
        path: '/settings',
        render: lazyComponent(() => import('./user/settings/RedirectToUserSettings'), 'RedirectToUserSettings'),
    },
    {
        path: '/user',
        render: lazyComponent(() => import('./user/settings/RedirectToUserPage'), 'RedirectToUserPage'),
    },
    {
        path: '/organizations',
        render: lazyComponent(() => import('./org/OrgsArea'), 'OrgsArea'),
    },
    {
        path: '/search',
        render: props => <SearchResults {...props} deployType={window.context.deployType} />,
        exact: true,
    },
    {
        path: '/site-admin/init',
        exact: true,
        render: props => <SiteInitPage {...props} context={window.context} />,
    },
    {
        path: '/site-admin',
        render: props => (
            <SiteAdminArea
                {...props}
                routes={props.siteAdminAreaRoutes}
                sideBarGroups={props.siteAdminSideBarGroups}
                overviewComponents={props.siteAdminOverviewComponents}
            />
        ),
    },
    {
        path: '/password-reset',
        render: lazyComponent(() => import('./auth/ResetPasswordPage'), 'ResetPasswordPage'),
        exact: true,
    },
    {
        path: '/api/console',
        render: lazyComponent(() => import('./api/ApiConsole'), 'ApiConsole'),
        exact: true,
    },
    {
        path: '/users/:username',
        render: lazyComponent(() => import('./user/area/UserArea'), 'UserArea'),
    },
    {
        path: '/survey/:score?',
        render: lazyComponent(() => import('./marketing/SurveyPage'), 'SurveyPage'),
    },
    {
        path: '/extensions',
        render: props => <ExtensionsArea {...props} routes={props.extensionsAreaRoutes} />,
    },
    {
        path: '/help',
        render: passThroughToServer,
    },
    {
        path: '/-/debug/*',
        render: passThroughToServer,
    },
    {
        path: '/insights',
        render: lazyComponent(() => import('./insights/InsightsRouter'), 'InsightsRouter'),
        condition: props =>
            !isErrorLike(props.settingsCascade.final) &&
            !!props.settingsCascade.final?.experimentalFeatures?.codeInsights &&
            props.settingsCascade.final['insights.displayLocation.insightsPage'] !== false,
    },
    {
        path: '/views',
        render: lazyComponent(() => import('./views/ViewsArea'), 'ViewsArea'),
    },
    {
        path: '/contexts',
        render: lazyComponent(() => import('./searchContexts/SearchContextsListPage'), 'SearchContextsListPage'),
        exact: true,
        condition: props =>
            !isErrorLike(props.settingsCascade.final) &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContext &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContextManagement,
    },
    {
        path: '/contexts/convert-version-contexts',
        render: lazyComponent(
            () => import('./searchContexts/ConvertVersionContextsPage'),
            'ConvertVersionContextsPage'
        ),
        exact: true,
        condition: props =>
            !isErrorLike(props.settingsCascade.final) &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContext &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContextManagement &&
            !!props.authenticatedUser?.siteAdmin,
    },
    {
        path: '/contexts/new',
        render: lazyComponent(() => import('./searchContexts/CreateSearchContextPage'), 'CreateSearchContextPage'),
        exact: true,
        condition: props =>
            !isErrorLike(props.settingsCascade.final) &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContext &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContextManagement,
    },
    {
        path: '/contexts/:id/edit',
        render: lazyComponent(() => import('./searchContexts/EditSearchContextPage'), 'EditSearchContextPage'),
        condition: props =>
            !isErrorLike(props.settingsCascade.final) &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContext &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContextManagement,
    },
    {
        path: '/contexts/:id',
        render: lazyComponent(() => import('./searchContexts/SearchContextPage'), 'SearchContextPage'),
        condition: props =>
            !isErrorLike(props.settingsCascade.final) &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContext &&
            !!props.settingsCascade.final?.experimentalFeatures?.showSearchContextManagement,
    },
    {
        path: '/refactor-python2-to-3',
        render: props => <RepogroupPage {...props} repogroupMetadata={python2To3Metadata} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        path: '/kubernetes',
        render: props => <RepogroupPage {...props} repogroupMetadata={kubernetes} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        path: '/stackstorm',
        render: props => <RepogroupPage {...props} repogroupMetadata={stackStorm} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        path: '/temporal',
        render: props => <RepogroupPage {...props} repogroupMetadata={temporal} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        path: '/golang',
        render: props => <RepogroupPage {...props} repogroupMetadata={golang} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        path: '/react-hooks',
        render: props => <RepogroupPage {...props} repogroupMetadata={reactHooks} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        path: '/android',
        render: props => <RepogroupPage {...props} repogroupMetadata={android} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        path: '/stanford',
        render: props => <RepogroupPage {...props} repogroupMetadata={stanford} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        path: '/cncf',
        render: props => <RepogroupPage {...props} repogroupMetadata={cncf} />,
        condition: ({ isSourcegraphDotCom }) => isSourcegraphDotCom,
    },
    {
        path: '/:repoRevAndRest+',
        render: lazyComponent(() => import('./repo/RepoContainer'), 'RepoContainer'),
    },
]
