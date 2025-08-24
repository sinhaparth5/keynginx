class Keynginx < Formula
  desc "SSL-enabled Nginx automation with Docker"
  homepage "https://github.com/sinhaparth5/keynginx"
  url "https://github.com/sinhaparth5/keynginx/releases/download/v1.0.0/keynginx-1.0.0-darwin-amd64.tar.gz"
  sha256 "YOUR_SHA256_HERE"
  license "MIT"

  depends_on "docker" => :recommended

  def install
    bin.install "keynginx"
  end

  test do
    system "#{bin}/keynginx", "version"
  end
end