/**
 * Copyright 2019, Z Lab Corporation. All rights reserved.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package vendordata

// https://docs.openstack.org/nova/latest/user/vendordata.html
type RequestData struct {
	// The ID of the project that owns this instance.
	ProjectID string `json:"project-id"`
	// The UUID of this instance.
	InstanceID string `json:"instance-id"`
	// The ID of the image used to boot this instance.
	ImageID string `json:"image-id"`
	// As specified by the user at boot time.
	UserData string `json:"user-data"`
	// The hostname of the instance.
	Hostname string `json:"hostname"`
	// As specified by the user at boot time.
	Metadata map[string]string `json:"metadata"`
}
