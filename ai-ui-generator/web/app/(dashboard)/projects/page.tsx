export default function ProjectsPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Projects</h1>
        <p className="text-gray-600">
          Manage and organize your UI generation projects.
        </p>
      </div>

      {/* Create New Project Button */}
      <div className="mb-6">
        <button className="bg-blue-600 text-white px-6 py-3 rounded-md hover:bg-blue-700 font-medium">
          Create New Project
        </button>
      </div>

      {/* Projects List */}
      <div className="bg-white rounded-lg border">
        <div className="px-6 py-4 border-b">
          <h2 className="text-lg font-medium text-gray-900">Recent Projects</h2>
        </div>
        <div className="p-6">
          <div className="text-center text-gray-500 py-8">
            <p>No projects yet. Create your first project to get started!</p>
          </div>
        </div>
      </div>
    </div>
  )
}
