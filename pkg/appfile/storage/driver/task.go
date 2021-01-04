package driver

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1alpha2"
	"github.com/oam-dev/kubevela/pkg/appfile"
	"github.com/oam-dev/kubevela/pkg/builtin"
	cmdutil "github.com/oam-dev/kubevela/pkg/commands/util"
	"github.com/oam-dev/kubevela/pkg/oam"
)

// Object generate v1alpha2.Application
func (app *Application) Object(ns string) (*v1alpha2.Application, []oam.Object, error) {
	servApp := new(v1alpha2.Application)
	servApp.SetNamespace(ns)
	servApp.SetName(app.Name)

	servApp.Spec.Components = []v1alpha2.ApplicationComponent{}
	for name, svc := range app.Services {
		comp := v1alpha2.ApplicationComponent{
			Name: name,
		}
		params := map[string]interface{}{}
		traits := []v1alpha2.ApplicationTrait{}
		for k, v := range svc {
			if k == "type" {
				comp.WorkloadType = v.(string)
				continue
			}
			if app.Tm.IsTrait(k) {
				trait := v1alpha2.ApplicationTrait{
					Name: k,
				}
				pts := &runtime.RawExtension{}
				jt, err := json.Marshal(v)
				if err != nil {
					return nil, nil, err
				}
				if err := pts.UnmarshalJSON(jt); err != nil {
					return nil, nil, err
				}
				trait.Properties = *pts
				traits = append(traits, trait)
				continue
			}
			params[k] = v
		}

		settings := &runtime.RawExtension{}
		pt, err := json.Marshal(params)
		if err != nil {
			return nil, nil, err
		}
		if err := settings.UnmarshalJSON(pt); err != nil {
			return nil, nil, err
		}
		comp.Settings = *settings
		if len(traits) > 0 {
			comp.Traits = traits
		}
		servApp.Spec.Components = append(servApp.Spec.Components, comp)
	}

	servApp.SetGroupVersionKind(v1alpha2.SchemeGroupVersion.WithKind("Application"))
	scopes := []oam.Object{}
	return servApp, scopes, nil
}

// InitTasks init tasks and generate new application
func (app *Application) InitTasks(io cmdutil.IOStreams) (*Application, error) {
	appFile := *app.AppFile
	newApp := Application{AppFile: &appFile, Tm: app.Tm}
	newApp.Services = map[string]appfile.Service{}
	for name, svc := range app.Services {
		newSvc, err := builtin.DoTasks(svc, io)
		if err != nil {
			return nil, err
		}
		newApp.Services[name] = newSvc
	}
	return &newApp, nil
}
