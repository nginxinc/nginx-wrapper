/*
 *  Copyright 2020 F5 Networks
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package version

import (
	"fmt"
	"github.com/nginxinc/nginx-wrapper/lib/osenv"
	"reflect"
	"strings"
	"testing"
)

func TestParsingNginxOpenSource(t *testing.T) {
	expected := NginxVersion{
		Full:    "nginx/1.17.9",
		Version: "1.17.9",
		Detail:  "",
		IsPlus:  false,
	}
	line := "nginx version: nginx/1.17.9"
	actual, err := parseVersion(line)

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected version doesn't match actual version:"+osenv.LineBreak+
			"expected: %v"+osenv.LineBreak+
			"actual:   %v", expected, actual)
	}
}

func TestParsingNginxPlus(t *testing.T) {
	expected := NginxVersion{
		Full:    "nginx/1.17.9 (nginx-plus-r21)",
		Version: "1.17.9",
		Detail:  "nginx-plus-r21",
		IsPlus:  true,
	}
	line := "nginx version: nginx/1.17.9 (nginx-plus-r21)"
	actual, err := parseVersion(line)

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected version doesn't match actual version:"+osenv.LineBreak+
			"expected: %v"+osenv.LineBreak+
			"actual:   %v", expected, actual)
	}
}

func TestParseConfigArgs(t *testing.T) {
	input := " --build=nginx-plus-r21 --prefix=/etc/nginx " +
		"--sbin-path=/usr/sbin/nginx --modules-path=/usr/lib/nginx/modules " +
		"--conf-path=/etc/nginx/nginx.conf " +
		"--error-log-path=/var/log/nginx/error.log " +
		"--http-log-path=/var/log/nginx/access.log " +
		"--pid-path=/var/run/nginx.pid --lock-path=/var/run/nginx.lock " +
		"--http-client-body-temp-path=/var/cache/nginx/client_temp " +
		"--http-proxy-temp-path=/var/cache/nginx/proxy_temp " +
		"--http-fastcgi-temp-path=/var/cache/nginx/fastcgi_temp " +
		"--http-uwsgi-temp-path=/var/cache/nginx/uwsgi_temp " +
		"--http-scgi-temp-path=/var/cache/nginx/scgi_temp --user=nginx " +
		"--group=nginx --with-compat --with-file-aio --with-threads " +
		"--with-http_addition_module --with-http_auth_jwt_module " +
		"--with-http_auth_request_module --with-http_dav_module " +
		"--with-http_f4f_module --with-http_flv_module " +
		"--with-http_gunzip_module " +
		"--with-http_gzip_static_module " +
		"--with-http_hls_module --with-http_mp4_module " +
		"--with-http_random_index_module --with-http_realip_module " +
		"--with-http_secure_link_module --with-http_session_log_module " +
		"--with-http_slice_module --with-http_ssl_module " +
		"--with-http_stub_status_module --with-http_sub_module " +
		"--with-http_v2_module --with-mail --with-mail_ssl_module " +
		"--with-stream --with-stream_realip_module --with-stream_ssl_module " +
		"--with-stream_ssl_preread_module " +
		"--with-cc-opt='-g -O2 -fdebug-prefix-map=/data/builder/debuild/nginx-plus-1.17.9/debian/debuild-base/nginx-plus-1.17.9=. -fstack-protector-strong -Wformat -Werror=format-security -Wp,-D_FORTIFY_SOURCE=2 -fPIC' --with-ld-opt='-Wl,-Bsymbolic-functions -Wl,-z,relro -Wl,-z,now -Wl,--as-needed -pie'"

	args := parseConfigureArgs(input)

	var builder strings.Builder
	for el := args.Front(); el != nil; el = el.Next() {
		key := fmt.Sprintf("%v", el.Key)
		val := fmt.Sprintf("%v", el.Value)
		builder.WriteString(" --")
		builder.WriteString(key)

		if val != "" {
			builder.WriteString("=")
			builder.WriteString(val)
		}
	}

	actual := builder.String()

	if input != actual {
		t.Errorf("parsed arguments didn't match expectation"+osenv.LineBreak+
			"expected: %s"+osenv.LineBreak+
			"actual:   %s", input, actual)
	}
}
